import type { Filter } from "../types";

interface Props {
  filter: Filter;
  onChange: (f: Filter) => void;
  counts: Record<Filter, number>;
}

const LABELS: Record<Filter, string> = {
  all: "All",
  active: "In progress",
  completed: "Completed",
};

/** Row of buttons to switch between All / In progress / Completed, showing how many tasks each has. */
export function FilterBar({ filter, onChange, counts }: Props) {
  return (
    <nav className="filters">
      {(Object.keys(LABELS) as Filter[]).map((f) => (
        <button
          key={f}
          className={f === filter ? "active" : ""}
          onClick={() => onChange(f)}
        >
          {LABELS[f]} ({counts[f]})
        </button>
      ))}
    </nav>
  );
}
