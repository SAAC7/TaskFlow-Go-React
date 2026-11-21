import type { Filter, Task, TaskPatch } from "../types";
import { TaskItem } from "./TaskItem";

interface Props {
  tasks: Task[];
  filter: Filter;
  onUpdate: (id: string, patch: TaskPatch) => Promise<void>;
  onDelete: (id: string) => Promise<void>;
}

const EMPTY_TEXT: Record<Filter, string> = {
  all: "Todavía no hay tareas. ¡Agregá la primera!",
  active: "No tenés tareas pendientes 🎉",
  completed: "Aún no completaste ninguna tarea.",
};

export function TaskList({ tasks, filter, onUpdate, onDelete }: Props) {
  if (tasks.length === 0) return <p className="empty">{EMPTY_TEXT[filter]}</p>;

  return (
    <ul className="task-list">
      {tasks.map((task) => (
        // key ayuda a React a identificar cada elemento; usá siempre un id estable.
        <TaskItem
          key={task.id}
          task={task}
          onUpdate={onUpdate}
          onDelete={onDelete}
        />
      ))}
    </ul>
  );
}
