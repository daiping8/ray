import logging
import os
from typing import Optional

import torch
import torch.distributed as dist

logger = logging.getLogger(__name__)

# 从环境变量设置日志级别（如果指定了的话）
dag_log_level = os.environ.get("RAY_DAG_LOG_LEVEL", "").upper()
if dag_log_level:
    try:
        level = getattr(logging, dag_log_level)
        logger.setLevel(level)
    except (AttributeError, ValueError):
        pass


class ScclCommunicator:
    def __init__(
        self,
        world_size: int,
        comm_id: tuple,
        rank: Optional[int],
    ):
        """Initialize an SCCL communicator for one device controlled by one
        process.

            Args:
                world_size: Total number of GPUs to be used.
                comm_id : The unique ID returned by :func:`get_unique_id`.
                rank (int): The rank of the GPU managed by the current process.

            Returns:
                ScclCommunicator: An ``ScclCommunicator`` instance.
        """
        logger.info(
            f"ScclCommunicator initialized with world_size: {world_size}, rank: {rank}"
        )
        assert dist.is_initialized()
        self.pg = dist.distributed_c10d._get_default_group()

    def send(self, tensor: torch.Tensor, dst: int, stream: int):
        logger.debug(f"send tensor to dst: {dst} ,tensor shape: {tensor.shape}")
        dist.send(tensor, dst, group=self.pg)

    def recv(self, tensor: torch.Tensor, src: int, stream: int):
        logger.debug(f"recv tensor from src: {src} ,tensor shape: {tensor.shape}")
        dist.recv(tensor, src, group=self.pg)

    def destroy(self):
        logger.info("destroy sccl communicator")
