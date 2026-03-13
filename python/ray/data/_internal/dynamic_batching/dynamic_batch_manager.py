import math
from collections import deque
from typing import Deque, Optional, Tuple


class DynamicBatchManager:
    """Simple latency-aware dynamic batch size controller.

    This manager keeps a sliding window of (batch_size, duration_s) samples and
    adjusts the recommended batch size range to keep the average latency close
    to the target.

    The policy is intentionally conservative and per-task only. It does not
    coordinate across workers or rely on any global state.
    """

    def __init__(
        self,
        initial_batch_size: int,
        *,
        target_latency_s: float = 5.0,
        min_batch_size: int = 1,
        max_batch_size: Optional[int] = None,
        window_size: int = 20,
        grow_factor: float = 1.5,
        shrink_factor: float = 0.7,
    ):
        if initial_batch_size <= 0:
            raise ValueError("initial_batch_size must be positive.")
        if min_batch_size <= 0:
            raise ValueError("min_batch_size must be positive.")
        if max_batch_size is not None and max_batch_size < min_batch_size:
            raise ValueError("max_batch_size must be >= min_batch_size.")

        self._target_latency_s = target_latency_s
        self._min_batch_size = min_batch_size
        self._max_batch_size = max_batch_size
        self._window_size = max(window_size, 1)
        self._grow_factor = max(grow_factor, 1.0)
        self._shrink_factor = min(shrink_factor, 1.0)

        self._samples: Deque[Tuple[int, float]] = deque(maxlen=self._window_size)

        # Current recommended bounds and batch size.
        self._lower_bound = initial_batch_size
        self._upper_bound = (
            max_batch_size if max_batch_size is not None else initial_batch_size
        )
        self._current_batch_size = initial_batch_size

    @property
    def current_bounds(self) -> Tuple[int, int]:
        return self._lower_bound, self._upper_bound

    @property
    def current_batch_size(self) -> int:
        return self._current_batch_size

    def record_execution_stats(self, batch_size: int, duration_s: float) -> None:
        """Record a single (batch_size, duration) observation."""
        if batch_size <= 0 or not math.isfinite(duration_s) or duration_s <= 0:
            return
        self._samples.append((batch_size, duration_s))
        self._current_batch_size = batch_size

    def calculate_batch_size(self) -> Tuple[int, int]:
        """Calculate a new (lower_bound, upper_bound) for batch size.

        This uses a simple heuristic:
        * If average duration is above target, shrink the current batch size.
        * If average duration is well below target, grow the current batch size.
        * Otherwise, keep the current bounds.
        """
        if not self._samples:
            return self._lower_bound, self._upper_bound

        # Compute average latency over recent samples.
        avg_duration = sum(d for _, d in self._samples) / len(self._samples)

        lower = self._lower_bound
        upper = self._upper_bound
        current = self._current_batch_size

        # Guard rails for current batch size.
        current = max(current, self._min_batch_size)
        if self._max_batch_size is not None:
            current = min(current, self._max_batch_size)

        # Small tolerance band to avoid oscillations.
        high_threshold = self._target_latency_s * 1.05
        low_threshold = self._target_latency_s * 0.95

        if avg_duration > high_threshold:
            # Too slow: shrink around current.
            new_upper = current
            new_lower = max(
                self._min_batch_size, int(max(current * self._shrink_factor, 1))
            )
        elif avg_duration < low_threshold:
            # Fast: we can try to grow.
            new_lower = current
            grown = int(max(current * self._grow_factor, current + 1))
            if self._max_batch_size is not None:
                grown = min(grown, self._max_batch_size)
            new_upper = max(grown, new_lower)
        else:
            # Within acceptable band; keep previous bounds.
            return lower, upper

        # Ensure monotonic within global [min, max] constraints.
        new_lower = max(new_lower, self._min_batch_size)
        if self._max_batch_size is not None:
            new_upper = min(new_upper, self._max_batch_size)

        # Avoid degenerate ranges.
        if new_upper < new_lower:
            new_upper = new_lower

        self._lower_bound = new_lower
        self._upper_bound = new_upper

        # Update current batch size to mid-point of the new range.
        self._current_batch_size = max(
            self._min_batch_size, (new_lower + new_upper) // 2
        )

        return self._lower_bound, self._upper_bound

