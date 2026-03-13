import time

from ray.data._internal.dynamic_batching import DynamicBatchManager


def test_dynamic_batch_manager_grow_and_shrink():
    manager = DynamicBatchManager(initial_batch_size=10, target_latency_s=1.0)

    # Fast batches should cause growth.
    for _ in range(5):
        manager.record_execution_stats(batch_size=10, duration_s=0.5)
    lower, upper = manager.calculate_batch_size()
    assert upper >= lower >= 1
    assert upper >= 10

    # Slow batches should cause shrink.
    for _ in range(5):
        manager.record_execution_stats(batch_size=upper, duration_s=2.0)
    new_lower, new_upper = manager.calculate_batch_size()
    assert new_upper <= upper
    assert new_lower >= 1


def test_dynamic_batch_manager_ignores_invalid_samples():
    manager = DynamicBatchManager(initial_batch_size=8, target_latency_s=1.0)

    manager.record_execution_stats(batch_size=0, duration_s=1.0)
    manager.record_execution_stats(batch_size=8, duration_s=-1.0)
    # No valid samples yet, bounds should remain unchanged.
    lower, upper = manager.calculate_batch_size()
    assert (lower, upper) == manager.current_bounds

