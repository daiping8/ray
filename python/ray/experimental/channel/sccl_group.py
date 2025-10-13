import logging
from types import ModuleType
from typing import TYPE_CHECKING, List, Optional, Tuple

import ray
from ray.exceptions import RayChannelError
from ray.experimental.channel.accelerator_context import AcceleratorContext
from ray.experimental.channel.communicator import Communicator, TorchTensorAllocator
from ray.experimental.util.types import ReduceOp

if TYPE_CHECKING:
    import torch
    import torch_br


# 该模块的日志记录器。应在使用Ray的程序入口点进行配置。
# Ray在入口/初始化点提供了默认配置。
logger = logging.getLogger(__name__)


class _ScclGroup(Communicator):
    """
    Represents an actor's SCCL communicator. This is the default SCCL communicator
    to be used in Compiled Graph if a custom communicator is not provided.

    This class is not thread-safe.
    """

    def __init__(
        self,
        world_size: int,
        comm_id: tuple,
        rank: Optional[int],
        actor_handles: List["ray.actor.ActorHandle"],
        supa_stream: Optional["torch_br.supa.Stream"],
        use_communication_streams: bool = False,
        ranks: Optional[List[int]] = None,
    ):
        """
        Initialize a SCCL communicator that can be used to communicate p2p with
        other GPU actors.

        This method blocks until the same call has been made on all other
        actors in the group, with the same arguments for world_size and
        comm_id.

        NOTE: A concurrent SCCL group can coexist with this one but using the
        two groups concurrently on different SUPA streams may cause deadlock.
        See SCCL documentation for details on concurrent communicator usage.

        If the user can guarantee that all involved actors execute the same ops
        in the same order, then the other SCCL group should use the given
        `supa_stream`, and there will not be a concurrency issue. Otherwise,
        the other stream needs to synchronize with the given `supa_stream`
        before and after it launches SCCL ops, e.g., at the beginning and end
        of a DAG task.

        Args:
            world_size: The number of participating actors/devices.
            comm_id: A unique communicator ID returned by
                SCCL's get_unique_id().
            rank: The rank of this actor. If None, then the caller is not a
                participant of the SCCL group.
            actor_handles: A list of actor handles, in rank order.
            supa_stream: A raw SUPA stream to dispatch SCCL ops to. If rank is
                specified, then this must be specified too.
            use_communication_streams: Whether to use dedicated send and recv
                streams for communication. If True, communication and computation
                can be overlapped to improve performance.
            ranks: A list of ranks for the actors which is assigned by the torch default process group.
        """
        self._world_size = world_size
        self._rank: Optional[int] = rank
        self.succl_util: Optional[ModuleType] = None
        self._actor_handles = actor_handles
        self._use_communication_streams = use_communication_streams
        self._ranks = ranks

        if use_communication_streams:
            raise NotImplementedError(
                "use_communication_streams is not implemented for SCCL group."
            )

        if rank is not None:
            assert ray.get_gpu_ids(), "SCCL actor has no GPUs assigned"
            assert supa_stream is not None, "SCCL actor must specify supa_stream"

            expected_rank = self.get_rank(ray.get_runtime_context().current_actor)
            assert (
                rank == expected_rank
            ), f"SCCL actor's rank {rank} does not match expected rank {expected_rank}"

            # 导入SCCL实用工具
            from ray.util.collective.collective_group import sccl_util

            self.succl_util = sccl_util
            # 创建SCCL通信器
            # 注意：这需要根据您的SCCL设置进行调整
            # 目前，我们将创建一个占位符
            self._comm = self.succl_util.ScclCommunicator(world_size, comm_id, rank)
        else:
            # Driver does not have a rank.
            self._comm = None

        self._supa_stream: Optional["torch_br.supa.Stream"] = None
        self._send_stream: Optional["torch_br.supa.Stream"] = None
        self._recv_stream: Optional["torch_br.supa.Stream"] = None
        if supa_stream is not None:
            assert rank is not None, "SCCL actor has no rank assigned"
            self._supa_stream = supa_stream

            if use_communication_streams:
                import torch_br

                # TODO(swang): 允许覆盖默认设备。
                device = AcceleratorContext.get().get_accelerator_devices()[0]

                self._send_stream = torch_br.supa.Stream(device=device.index)
                self._recv_stream = torch_br.supa.Stream(device=device.index)
            else:
                self._send_stream = self._supa_stream
                self._recv_stream = self._supa_stream

        self._closed = False

    def initialize(self, rank: int) -> None:
        # No additional initialization is needed.
        pass

    def get_actor_handles(self) -> List["ray.actor.ActorHandle"]:
        return self._actor_handles

    def get_rank(self, actor: ray.actor.ActorHandle) -> int:
        """
        Return the given actor's rank in the SCCL communicator.

        Args:
            actor: The actor handle to look up.
        """
        actor_ids = [a._ray_actor_id for a in self._actor_handles]
        try:
            rank_index = actor_ids.index(actor._ray_actor_id)
        except ValueError:
            raise ValueError("Actor is not in the SCCL group.")
        if self._ranks is not None:
            return self._ranks[rank_index]
        else:
            return rank_index

    def get_self_rank(self) -> Optional[int]:
        """
        Return this actor's rank.
        """
        return self._rank

    def get_world_size(self) -> int:
        """
        Return the number of ranks in the SCCL communicator.
        """
        return self._world_size

    def send(self, buf: "torch.Tensor", peer_rank: int) -> None:
        """
        Send a torch.Tensor to a peer.

        This returns when the send kernel has been queued, but the kernel may
        not have completed. Therefore, the caller should ensure that there are
        no concurrent writes to the sent `buf` until the send has finished.
        That is, either all writes should be submitted on the current stream
        (self._supa_stream) or, if on a different stream, that stream should
        synchronize with the current stream.

        Args:
            buf: The torch.Tensor to send. It should already be on this
                actor's default device.
            peer_rank: The rank of the actor to send to.
        """
        if self._closed:
            raise RayChannelError("SCCL group has been destroyed.")

        if self._use_communication_streams:
            # 我们观察到如果所有接收/计算/发送操作都在GPU上运行，
            # 由于没有同步，CPU执行循环可能会远超GPU操作并导致运行时失败。
            # 为了避免这种情况，我们在发送流上进行同步。
            # TODO(rui): 寻找更好的方法
            self._send_stream.synchronize()

        # TODO(swang): 处理发送/接收异步SCCL错误，如网络故障。
        self._comm.send(buf, peer_rank, self._send_stream.supa_stream)

    def recv(
        self,
        shape: Tuple[int],
        dtype: "torch.dtype",
        peer_rank: int,
        allocator=Optional[TorchTensorAllocator],
    ) -> "torch.Tensor":
        """
        Receive a torch.Tensor from a peer and synchronize the current stream.

        After this call returns, the receive buffer is safe to read from from
        any stream. An RayChannelError will be raised if an error occurred (e.g.,
        remote actor died), and the buffer is not safe to read.

        Args:
            shape: The shape of the tensor to receive.
            dtype: The dtype of the tensor to receive.
            peer_rank: The rank of the actor to receive from.
            allocator: A function to allocate the tensor to receive into.
        """
        if self._closed:
            raise RayChannelError("SCCL group has been destroyed.")
        assert allocator is not None, "SCCL group requires a tensor allocator"
        buf = allocator(shape, dtype)

        if self._use_communication_streams:
            # 我们观察到如果所有接收/计算/发送操作都在GPU上运行，
            # 由于没有同步，CPU执行循环可能会远超GPU操作并导致运行时失败。
            # 为了避免这种情况，我们在接收流上进行同步。
            # TODO(rui): 寻找更好的方法
            self._recv_stream.synchronize()

            self._comm.recv(buf, peer_rank, self._recv_stream.supa_stream)
        else:
            self._comm.recv(buf, peer_rank, self._recv_stream.supa_stream)

            # 如果SCCL操作被中止，缓冲区值将是未定义的。因此，我们需要在此处
            # 进行同步并检查通道是否仍然打开，以确保接收缓冲区有效。
            # TODO(swang): 避免SUPA同步。

            self._supa_stream.synchronize()

        if self._closed:
            raise RayChannelError("SCCL group has been destroyed.")
        return buf

    # TODO：实现SCCL的集合通信
    def allreduce(
        self,
        send_buf: "torch.Tensor",
        recv_buf: "torch.Tensor",
        op: ReduceOp = ReduceOp.SUM,
    ):
        raise NotImplementedError("allreduce is not implemented for SCCL group.")

    def allgather(
        self,
        send_buf: "torch.Tensor",
        recv_buf: "torch.Tensor",
    ):
        raise NotImplementedError("Allgather is not implemented for SCCL group.")

    def reducescatter(
        self,
        send_buf: "torch.Tensor",
        recv_buf: "torch.Tensor",
        op: ReduceOp = ReduceOp.SUM,
    ):
        raise NotImplementedError("Reducescatter is not implemented for SCCL group.")

    @property
    def recv_stream(self):
        import torch_br

        return torch_br.supa.stream(self._recv_stream)

    @property
    def send_stream(self):
        import torch_br

        return torch_br.supa.stream(self._send_stream)

    def destroy(self) -> None:
        """
        销毁SCCL组。
        """
        if self._closed:
            return

        self._closed = True

        if self._comm is not None:
            logger.info(
                "正在销毁actor上的SCCL组："
                f"{ray.get_runtime_context().current_actor}"
            )
            # 在设置_closed标志*之后*中止。这确保了那些因远程peer而阻塞的SCCL
            # 操作在退出中止时会看到_closed标志为True。

            # self._comm.abort()
            self._comm.destroy()

    # transport_name统一为"accelerator"
    def get_transport_name(self) -> str:
        return "accelerator"

    @classmethod
    def generate_communicator_id(cls) -> str:
        """
        Generate a unique identifier for the SCCL communicator.

        Raises:
            NotImplementedError: This method is not yet implemented for SCCL.

        Returns:
            str: A unique identifier that can be used to create a SCCL communicator group.
        """
        raise NotImplementedError(
            "generate_communicator_id is not implemented for SCCL group."
        )
