import { useState, type FormEvent } from "react";
import { getErrorMessage } from "../api";
import type { TaskInput } from "../types";

interface Props {
  onAdd: (input: TaskInput) => Promise<void>;
}

export function TaskForm({ onAdd }: Props) {
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

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
        placeholder="¿Qué hay que hacer?"
        maxLength={120}
        disabled={submitting}
        aria-label="Título"
      />
      <textarea
        value={description}
        onChange={(e) => setDescription(e.target.value)}
        placeholder="Descripción (opcional)"
        maxLength={1000}
        rows={2}
        disabled={submitting}
        aria-label="Descripción"
      />
      <button type="submit" disabled={submitting}>
        {submitting ? "Guardando…" : "Agregar tarea"}
      </button>
      {error && <p className="error">{error}</p>}
    </form>
  );
}
