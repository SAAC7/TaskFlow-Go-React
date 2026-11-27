import { useState } from 'react'
import { getErrorMessage } from '../api'
import type { Task, TaskPatch } from '../types'

interface Props {
  task: Task
  onUpdate: (id: string, patch: TaskPatch) => Promise<void>
  onDelete: (id: string) => Promise<void>
}

/** One task of the list: checkbox to complete it, inline editing of title and description, and a delete button. */
export function TaskItem({ task, onUpdate, onDelete }: Props) {
  const [editing, setEditing] = useState(false)
  const [draftTitle, setDraftTitle] = useState(task.title)
  const [draftDescription, setDraftDescription] = useState(task.description)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)

  /** Wraps any action: marks the item as busy, captures the error and shows it. */
  async function run(action: () => Promise<void>) {
    setBusy(true)
    setError(null)
    try {
      await action()
    } catch (err) {
      setError(getErrorMessage(err))
    } finally {
      setBusy(false)
    }
  }

  /** Switches to edit mode, starting from the current values of the task. */
  function startEditing() {
    setDraftTitle(task.title)
    setDraftDescription(task.description)
    setError(null)
    setEditing(true)
  }

  /** Leaves edit mode and discards the unsaved changes. */
  function cancelEditing() {
    setEditing(false)
    setError(null)
  }

  /** Marks the task as completed, or back to pending, with a PATCH request. */
  const toggle = () => run(() => onUpdate(task.id, { completed: !task.completed }))

  /** Marks the task as completed, or back to pending, with a PATCH request. */
  const remove = () => run(() => onDelete(task.id))

  /** Saves the edited title and description with a PATCH request (the title is required). */
  const save = () =>
    run(async () => {
      if (!draftTitle.trim()) {
        throw new Error('El título es obligatorio')
      }
      await onUpdate(task.id, { title: draftTitle, description: draftDescription })
      setEditing(false)
    })

  return (
    <li className={`task ${task.completed ? 'done' : ''}`}>
      <input
        type="checkbox"
        checked={task.completed}
        onChange={toggle}
        disabled={busy || editing}
        aria-label="Marcar como completada"
      />

      <div className="task-body">
        {editing ? (
          <div className="task-edit">
            <input
              value={draftTitle}
              onChange={(e) => setDraftTitle(e.target.value)}
              maxLength={120}
              autoFocus
              disabled={busy}
              aria-label="Editar título"
            />
            <textarea
              value={draftDescription}
              onChange={(e) => setDraftDescription(e.target.value)}
              maxLength={1000}
              rows={3}
              disabled={busy}
              placeholder="Descripción (opcional)"
              aria-label="Editar descripción"
            />
          </div>
        ) : (
          <>
            <span className="task-title">{task.title}</span>
            {task.description && <p className="task-desc">{task.description}</p>}
            <small>{new Date(task.createdAt).toLocaleDateString('es-GT')}</small>
          </>
        )}
        {error && <p className="error">{error}</p>}
      </div>

      <div className="task-actions">
        {editing ? (
          <>
            <button onClick={save} disabled={busy}>Save</button>
            <button onClick={cancelEditing} disabled={busy}>Cancel</button>
          </>
        ) : (
          <>
            <button onClick={startEditing} disabled={busy}>Edit</button>
            <button className="danger" onClick={remove} disabled={busy}>Delete</button>
          </>
        )}
      </div>
    </li>
  )
}