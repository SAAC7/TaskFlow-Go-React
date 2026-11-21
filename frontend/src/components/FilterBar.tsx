import type { Filter } from "../types";

interface Props {
  filter: Filter;
  onChange: (f: Filter) => void;
  counts: Record<Filter, number>;
}

const LABELS: Record<Filter, string> = {
  all: "Todas",
  active: "Pendientes",
  completed: "Completadas",
};

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
