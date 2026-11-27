import { useEffect, useState } from 'react'
import { api, getErrorMessage } from '../api'
import type { Task, TaskInput, TaskPatch } from '../types'


/**
* Custom hook that owns the task state (list, loading and error) and the
* functions that talk to the API. The components only render what it returns.
*/
export function useTasks() {
  const [tasks, setTasks] = useState<Task[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [attempt, setAttempt] = useState(0) // subirlo vuelve a ejecutar el efecto

  useEffect(() => {
    const controller = new AbortController()

    api
      .list(controller.signal)
      .then((data) => {
        setTasks(data)
        setError(null)
      })
      .catch((err) => {
        if (controller.signal.aborted) return
        setError(getErrorMessage(err))
      })
      .finally(() => {
        if (!controller.signal.aborted) setLoading(false)
      })

    // Cleanup: If the component is unmounted, we cancel the in-flight request.
    return () => controller.abort()
  }, [attempt])

  /** Loads the tasks again (used by the "Try Again" button after an error). */
  function reload() {
    setLoading(true)
    setError(null)
    setAttempt((a) => a + 1)
  }

  // The mutations allow the error to “escalate” (throw) so that the component
  // that called them (form, card) can display it where appropriate.
  async function addTask(input: TaskInput) {
    const created = await api.create(input)
    setTasks((prev) => [created, ...prev]) // el backend ordena "más nuevas primero"
  }


  /** Updates a task through the API (PATCH) and replaces it in the list with the version the server returns. */
  async function updateTask(id: string, patch: TaskPatch) {
    const updated = await api.patch(id, patch)
    setTasks((prev) => prev.map((t) => (t.id === id ? updated : t)))
  }


  /** Deletes a task through the API and removes it from the list. */
  async function deleteTask(id: string) {
    await api.remove(id)
    setTasks((prev) => prev.filter((t) => t.id !== id))
  }

  return { tasks, loading, error, reload, addTask, updateTask, deleteTask }
}