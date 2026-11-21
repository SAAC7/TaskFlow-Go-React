// Mirror of models.Task (Go). Dates are provided as ISO 8601 strings.
export interface Task {
  id: string
  title: string
  description: string
  completed: boolean
  createdAt: string
  updatedAt: string
}

// What we send when creating (models.TaskInput)
export interface TaskInput {
  title: string
  description: string
  completed?: boolean
}

// What we send in the PATCH request: all optional (models.TaskPatch)
export type TaskPatch = Partial<Pick<Task, 'title' | 'description' | 'completed'>>

// Your API's custom error format (response.go)
export interface ApiErrorBody {
  error: {
    code: string
    message: string
    details?: Record<string, string>
  }
}

export type Filter = 'all' | 'active' | 'completed'