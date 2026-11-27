import { useState } from "react";
import { useTasks } from "./hooks/useTasks";
import { TaskForm } from "./components/TaskForm";
import { TaskList } from "./components/TaskList";
import { FilterBar } from "./components/FilterBar";
import type { Filter } from "./types";

/**
 * Root component: loads the tasks with the useTasks hook and composes the form,
 * the filter bar and the list, showing the loading and error states.
 */
function App() {
  const { tasks, loading, error, reload, addTask, updateTask, deleteTask } =
    useTasks();
  const [filter, setFilter] = useState<Filter>("all");

  // Valores derivados: se calculan en cada render, no necesitan estado propio.
  const counts: Record<Filter, number> = {
    all: tasks.length,
    active: tasks.filter((t) => !t.completed).length,
    completed: tasks.filter((t) => t.completed).length,
  };
  const visible = tasks.filter((t) =>
    filter === "all" ? true : filter === "active" ? !t.completed : t.completed,
  );

  return (
    <main className="app">
      <h1>TaskFlow</h1>
      <TaskForm onAdd={addTask} />

      {loading && <p className="status">Loading tasks…</p>}

      {error && (
        <div className="status error">
          <p>{error}</p>
          <button onClick={reload}>Try Again</button>
        </div>
      )}

      {!loading && !error && (
        <>
          <FilterBar filter={filter} onChange={setFilter} counts={counts} />
          <TaskList
            tasks={visible}
            filter={filter}
            onUpdate={updateTask}
            onDelete={deleteTask}
          />
        </>
      )}
    </main>
  );
}

export default App;
