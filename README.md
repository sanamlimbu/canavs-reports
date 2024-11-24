# Canvas Report

A web service that interacts with Canvas LMS API to provide information about courses, assignments, enrollments, etc. as comprehensive reports.

## Features

- Fetch ungraded assignments for a specific course, organised by section.
- Retrieve student enrollments and assignments result.

## Prerequisites

- Canvas LMS admin account and admin's API access token

## Steps to Run

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

## Authentication

Currently, there is no authentication implemented. However, `withAuth` middleware is available and can be used if authentication is required.

Supabase can be used for user authentication, handling login and token generation seamlessly. The client can then include the token as a bearer token in API requests to authenticate users.

# Deployment

The provided Terraform files deploy a Go binary as an AWS Lambda function behind an API Gateway. Modify the Terraform configuration and make any necessary adjustments to meet the requirements.
