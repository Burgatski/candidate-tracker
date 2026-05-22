# Remotely challenge

This challenge consists of building a very simple app for a managing candidates. There is no frontend provided, but a small in-memory backend in Go is provided. Part of the task will be changing that backend to not be in-memory anymore.

### What if I do not know Go?

Go is fairly easy to read, but if you don't feel confident with Go you can roll your own in your preferred stack (but please, include instructions on how to run it).

### Setup

You can set up the project with the starter or toolkit of your choosing, as long as the project works by running `npm start` on the delivered project.

The project **must** use **plain React** (no Next.js, Remix or any other framework). Any other stack choice, as well as the design, is left to the candidate. Do not overspend time in making it pretty, the important thing is that it works.

### Delivered project

The delivered project should be a git repository. Ideally, the changes will be added incrementally in different commits.

## Challenge

The challenge is building a simple application to manage a list of candidates. Changes will be added iteratively through the following points. You do not need to complete all of these, but the more you complete, the better.

Adding tests is not mandatory, but will be a good bonus, if added.

**1. Candidate list**

The index of the app should be a list of candidates where each row shows the picture, the name, the email and the skills of the candidate. This list should be paginated, so the user can browse through all candidates.

**2. Candidate detail**

From the list on the index, an user should be able to go to the detail of the candidate in a separate page. This page should have its own route. In this page all the candidate data should be displayed.

**3. Create candidates**

On the index of the app there should be a "Create candidate" button. Upon clicking on such button, the user should see a form (whether it's a modal or a separate page is left to the candidate) with the following fields:

- First name (required)
- Last name (required)
- Email (required, must validate it's a valid email)
- Phone (required)
- Picture (must be able to select a file, required)
- Skills (multiple, optional)

When the form is submitted, if the response is ok, the user should be redirected to the details page to see the newly created candidate.

**4. Edit candidates**

When viewing and individual candidate profile there should be a button to edit the candidate data. Upon clicking on such button, the user should see a form (whether it's a modal or a separate page is left to the candidate) with the following fields prefilled with existing data:

- First name (required)
- Last name (required)
- Email (required, must validate it's a valid email)
- Phone (required)
- Picture (must be able to select a file, required)
- Skills (multiple, optional)

When the form is submitted, if the response is ok, the user should be redirected to the details page to see the updated candidate.

**5. Persistence**

As you may have noticed by now, the data of the candidates is stored in memory. You should change the backend to persist it on a SQL database of your choice. Feel free to make all the changes you want in logic, structure and architecture during this step.

**6. (BONUS) Delete candidates**

You will need to implement a `DELETE /candidates/{id}` endpoint in the server and then add a delete button (either in the list or the detail) to delete candidates. User should be asked to confirm the deletion before doing so.

## Backend

The backend is a very simple REST API for retrieving, updating and creating candidates.

### Running the backend

You can either run it with go

```
cd server && go run .
```

or with Docker using the provided dockerfile

```
cd server
docker build -t remotely-server .
docker run -d -p 8080:8080 --name remotely-server remotely-server
```

The backend has CORS enabled and will run on `localhost:8080`.

### Endpoints

#### GET /candidates

Returns up to 25 candidates depending on the requested page. Accepts a `page` query parameter that must be a number.

**Example response**:

```json
{
  "status": 200,
  "data": {
    "total": 80,
    "candidates": [
      {
        "id": 1,
        "first_name": "John",
        "last_name": "Doe",
        "picture": "<data uri>",
        "phone": "<phone>",
        "email": "<email>",
        "skills": ["Java", "Kotlin"]
      }
    ],
    "pagination": {
      "per_page": 25,
      "page": 1,
      "total_pages": 4
    }
  }
}
```

#### GET /candidates/{id}

Returns the requested candidate by ID.

**Example response**:

```
{
	"status": 200,
	"data": {
		"id": 1,
		"first_name": "John",
		"last_name": "Doe",
		"picture": "<data uri>",
		"phone": "<phone>",
		"email": "<email>",
		"skills": ["Java", "Kotlin"]
	}
}
```

**Example response if not found**:

```
{
	"status": 404,
	"errors": ["not found"]
}
```

#### POST /candidates

Creates a new candidate on the database. **NOTE**: changes are ephemeral, if the server restarts, the changes are lost.

Expects a JSON body with the following shape:

```
{
	"first_name": "John",
	"last_name": "Doe",
	"picture": "<data uri>",
	"phone": "<phone>",
	"email": "<email>",
	"skills": ["Java", "Kotlin"]
}
```

**Example response**:

```
{
	"status": 201,
	"data": {
		"id": 1,
		"first_name": "John",
		"last_name": "Doe",
		"picture": "<data uri>",
		"phone": "<phone>",
		"email": "<email>",
		"skills": ["Java", "Kotlin"]
	}
}
```

**Example response if there are errors**:

```
{
	"status": 400,
	"errors": ["first_name is not valid", "last_name is not valid]
}
```

#### PATCH /candidates/{id}

Updates the candidate with the given id. **NOTE**: changes are ephemeral, if the server restarts, the changes are lost.

Expects a JSON body with the following shape:

```
{
	"first_name": "John",
	"last_name": "Doe",
	"picture": "<data uri>",
	"phone": "<phone>",
	"email": "<email>",
	"skills": ["Java", "Kotlin"]
}
```

**Example response**:

```
{
	"status": 200,
	"data": {
		"id": 1,
		"first_name": "John",
		"last_name": "Doe",
		"picture": "<data uri>",
		"phone": "<phone>",
		"email": "<email>",
		"skills": ["Java", "Kotlin"]
	}
}
```

**Example response if there are errors**:

```
{
	"status": 400,
	"errors": ["first_name is not valid", "last_name is not valid]
}
```
