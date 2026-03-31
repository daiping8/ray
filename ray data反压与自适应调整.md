# Ray Data 反压与自适应调整

## 1. 简要结论

- **有“自适应”**：Ray Data 在下游出现反压时，会**自动限流/停发任务**，并限制从运行中的任务再读取多少输出。
- **不会“调 batch 大小”**：这些自适应体现在**能否再给算子喂新输入、能从运行中任务读多少字节**，而**不会自动修改 `map_batches` 的 `batch_size` 等参数**。

下文分别说明：反压如何检测与触发，以及“自适应”具体调整了什么。

---

## 2. 反压如何被检测和触发？

反压逻辑位于：

- `ray.data._internal.execution.backpressure_policy.*`
- `ray.data._internal.execution.resource_manager.ResourceManager`
- `ray.data._internal.execution.streaming_executor_state.select_operator_to_run()`

主要策略有两种：

### 2.1 DownstreamCapacityBackpressurePolicy

**文件**：`python/ray/data/_internal/execution/backpressure_policy/downstream_capacity_backpressure_policy.py`

**作用**：当下游队列堆积、处理能力不足时，对上游算子施加反压。

**核心指标**：

- **队列大小**（`queue_size_bytes`）：
  - 当前算子输出队列 + 下游 ineligible 算子的对象存储占用。
  - 由 `_get_queue_size_bytes(op)` 计算。
- **下游容量**（`downstream_capacity_size_bytes`）：
  - 下游 eligible 算子的 pending task inputs 之和。
  - 若无下游算子，则用外部消费者缓冲字节数（`external_consumer_bytes`）。
  - 由 `_get_downstream_capacity_size_bytes(op)` 计算。

**计算逻辑**：

- `queue_ratio = queue_size_bytes / downstream_capacity_size_bytes`
- 若下游容量为 0，则 `queue_ratio = 0`，不施加反压。
- 此外还会检查**对象存储预算使用率** `utilized_budget_fraction` 是否超过阈值  
  `RAY_DATA_DOWNSTREAM_CAPACITY_OBJECT_STORE_BUDGET_UTIL_THRESHOLD`（默认 0.9）。

**何时触发反压**：

- `utilized_budget_fraction > OBJECT_STORE_BUDGET_UTIL_THRESHOLD`（对象存储已较满），且
- `queue_ratio > downstream_capacity_backpressure_ratio`（来自 `DataContext.downstream_capacity_backpressure_ratio`）。

**反压时的行为**：

- `can_add_input(op)` → 返回 `False`：不再给该算子派发新的输入 bundle。
- `max_task_output_bytes_to_read(op)` → 返回 `0`：不再从该算子的运行中任务拉取新的输出块。

### 2.2 ResourceBudgetBackpressurePolicy

**文件**：`python/ray/data/_internal/execution/backpressure_policy/resource_budget_backpressure_policy.py`

**作用**：基于 `ResourceManager` 的资源预算（CPU、内存、对象存储）进行反压。

**决策逻辑**：

- `can_add_input(op)` 委托给 `OpResourceAllocator.can_submit_new_task(op)`：
  - 若预算不足，则不允许为该算子提交新任务。
- `max_task_output_bytes_to_read(op)` 委托给 `ResourceManager.max_task_output_bytes_to_read(op)`：
  - 根据当前对象存储使用情况，限制从任务输出中读取的字节数。

---

## 3. 自适应如何调整？

`StreamingExecutor` 在每轮 `_scheduling_loop_step` 中会：

1. 调用 `resource_manager.update_usages()` 更新各算子资源占用。
2. 调用 `select_operator_to_run()`，其中会使用 backpressure policy 的 `can_add_input` / `max_task_output_bytes_to_read` 结果。
3. 根据这些结果决定：
   - **哪些算子可以继续派发新任务**；
   - **从哪些算子的运行中任务中继续读取输出**。

**会被调整的**：

- **任务级调度与拉取速率**：
  - 若某上游算子的下游“吃不消”，该算子会被标记为 backpressured。
  - 不再为其派发新的输入 bundle（任务数不再增加，或逐渐跑完）。
  - 不再从其运行中的任务中拉取更多输出块（`max_task_output_bytes_to_read = 0`），让下游先消化已有数据。
- **资源预算分配**（若启用 `OpResourceAllocator`）：
  - `ResourceManager` 会根据全局资源与各算子占用，动态调整 per-op 预算。
  - `ResourceBudgetBackpressurePolicy` 据此决定是否允许提交新任务、允许读取多少输出。

**不会被调整的**：

- **不会自动修改 `map_batches` 的 `batch_size`**。
- 不会自动修改 UDF、`iter_batches` 的 `batch_size` 或 `local_shuffle_buffer_size`。
- 反压机制停留在 **限流/停发 + 限制拉取** 层面，不涉及动态批大小控制。

---

## 4. 总结

- **当前 Ray Data 的行为**：  
  下游压力大 → 调度器认为某算子处于反压状态 → 不再为其派发新输入、并限制从其任务中读取输出 → 上游“自然慢下来”，保护下游和对象存储。

- **尚未实现的行为**：  
  下游压力大 → 自动减小该算子的 batch size，以降低单批体积，从而减少队列堆积和内存峰值。这一部分正是文档《ray data动态批处理.md》第 9 章中设计要实现的扩展能力。
