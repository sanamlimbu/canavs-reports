# Canvas Report

A web service that interacts with Canvas LMS API to provide information about courses, assignments, and enrollments as comprehensive reports.

---

## Features

- Fetch ungraded assignments for a specific course, organized by section.
- Retrieve student enrollments and assignments result.

---

### Prerequisites

- Canvas LMS admin account and admin's API access token

---

### Steps to Run

1. Clone the repository.

   ```bash
   git clone https://github.com/sanamlimbu/canvas-report.git
   cd canvas-report

   ```

2. Set up environment variables.

   ```bash
   export CANVAS_BASE_URL=<your_canvas_base_url>
   export CANVAS_ACCESS_TOKEN=<your_canvas_access_token>
   export CANVAS_PAGE_SIZE=100
   ```

3. Build and run the application.
   ```bash
   go run cmd/server/main.go
   ```
