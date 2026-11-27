import { useState, type FormEvent } from "react";
import { getErrorMessage } from "../api";
import type { TaskInput } from "../types";

interface Props {
  onAdd: (input: TaskInput) => Promise<void>;
}

/** Form to create a task: checks the title, sends it through onAdd and shows any error. */
export function TaskForm({ onAdd }: Props) {
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  /** Runs when the form is submitted: validates the title, creates the task and clears the form on success. */
  async function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault(); // evita que el navegador recargue la página
    if (!title.trim()) {
      setError("El título es obligatorio");
      return;
    }
    setSubmitting(true);
    setError(null);
    try {
      await onAdd({ title, description });
      setTitle("");
      setDescription("");
    } catch (err) {
      setError(getErrorMessage(err));
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <form className="task-form" onSubmit={handleSubmit}>
      <input
        value={title}
        onChange={(e) => setTitle(e.target.value)}
        placeholder="What should we do?"
        maxLength={120}
        disabled={submitting}
        aria-label="Title"
      />
      <textarea
        value={description}
        onChange={(e) => setDescription(e.target.value)}
        placeholder="Description (optional)"
        maxLength={1000}
        rows={2}
        disabled={submitting}
        aria-label="Description"
      />
      <button type="submit" disabled={submitting}>
        {submitting ? "Saving…" : "Add Task"}
      </button>
      {error && <p className="error">{error}</p>}
    </form>
  );
}
