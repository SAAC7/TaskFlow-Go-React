# Overview

As a software engineer expanding my full-stack development capabilities, I am building **TaskFlow Web** to deepen my practical experience designing clean backend APIs in Go alongside reactive, strongly typed user interfaces in React and TypeScript. Rather than relying on pre-built application templates, I am writing both the server and client code from scratch to master end-to-end data flow, CORS configuration, and real-time state synchronization.

**TaskFlow Web** is a full-stack task management web application powered by a Go REST API (using Go structs and `net/http` endpoints for `GET`, `POST`, and `DELETE` operations) and a React + TypeScript frontend initialized with Vite. To start the test servers locally on your machine, run the Go backend server first (for example, `go run .` on `http://localhost:8080`) and start the React development server (`npm install` and `npm run dev` on `http://localhost:5173`), then open `http://localhost:5173` in your web browser to view the main page of the application.

My purpose for writing this software is to strengthen my command of Go HTTP handlers and data modeling, TypeScript interfaces, React Hooks (`useState` and `useEffect`), and asynchronous `fetch` requests while maintaining clear documentation and a repeatable full-stack workflow.

[Software Demo Video](http://youtube.link.goes.here)

# Web Pages

- **Main Task Dashboard (`/`)**: The primary page of the web application where users view, create, and delete tasks. Transitions happen smoothly across dynamic UI states—such as initial loading, populated task list, empty list state, and instant list updates after creating or deleting tasks—without full browser reloads.
  - **Dynamically Created Elements**:
    - **Task List and Task Cards**: Rendered dynamically from the JSON task array fetched via `GET` from the Go API, displaying each task's details and a delete action (`DELETE`).
    - **Task Creation Form**: Captures user input, validates against TypeScript interfaces, dispatches `POST` requests to the Go backend, and updates the React component state (`useState` and `useEffect`) in real time.

# Development Environment

- **Tools Used**: Visual Studio Code, Git, GitHub, Dev Containers, Browser DevTools (Network tab), and cURL / Postman for API testing.
- **Programming Language & Libraries**: Go (Golang) standard library (`net/http`, `encoding/json`, and Go structs) for the backend REST API and CORS handling, paired with React, TypeScript, Vite, React Hooks (`useState`, `useEffect`), the Fetch API, and CSS / Tailwind for the frontend interface.

# Useful Websites

- [The Go Programming Language Documentation](https://go.dev/doc/)
- [Go net/http Package Reference](https://pkg.go.dev/net/http)
- [React Official Documentation](https://react.dev/)
- [TypeScript Handbook](https://www.typescriptlang.org/docs/)
- [Vite Documentation](https://vite.dev/guide/)
- [MDN Web Docs - Fetch API](https://developer.mozilla.org/en-US/docs/Web/API/Fetch_API/Using_Fetch)

# Future Work

- Add task editing (`PUT` / `PATCH` endpoints) along with task completion toggles and filtering options (all, active, completed).
- Replace temporary/in-memory storage with a persistent database (such as SQLite or PostgreSQL) in the Go backend.
- Add automated unit and integration tests for Go HTTP routes and React components, plus richer error handling and loading states in the UI.
