import pyarrow as pa

from ray.data._internal.batcher import DynamicBatcher


def test_dynamic_batcher_update_bounds_changes_batch_size():
    table = pa.table({"x": list(range(100))})

    batcher = DynamicBatcher(batch_size=10, ensure_copy=False)
    batcher.add(table)
    batcher.done_adding()

    # Initial batch size 10.
    assert batcher.has_batch()
    batch = batcher.next_batch()
    assert len(batch) == 10

    # Update to a larger batch size.
    batcher.update_bounds(20, 40)
    assert batcher.has_batch()
    batch = batcher.next_batch()
    assert len(batch) in (20, 30, 40)  # depending on remaining rows

    # Invalid bounds should raise.
    try:
        batcher.update_bounds(0, 10)
        assert False, "Expected ValueError for non-positive bounds"
    except ValueError:
        pass

