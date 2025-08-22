# Essay Feature Documentation

## Overview

The Essay feature allows users to manage essays associated with their college applications. Each essay is linked to a college application and tracks the essay's progress through various stages.

## Models

### Essay Model

- **Title**: The title/topic of the essay
- **CollegeApplicationID**: Reference to the college application this essay belongs to
- **Status**: Current status of the essay (not_started, in_progress, not_submitted, review_needed, submitted, completed)
- **Prompt**: The essay prompt or question
- **Content**: The actual essay content
- **AIScore**: AI-generated score (0-10 scale)
- **AIFeedbacks**: Array of AI-generated feedback comments
- **AISuggestion**: AI-generated suggestions for improvement

### Essay Status Enum

- `not_started`: Essay hasn't been started yet
- `in_progress`: Essay is being written
- `not_submitted`: Essay is complete but not submitted
- `review_needed`: Essay needs review/revision
- `submitted`: Essay has been submitted
- `completed`: Essay process is fully completed

## API Endpoints

### Create Essay

**POST** `/essays`

**Request Body:**

```json
{
  "title": "Personal Statement",
  "college_application_id": "64a1b2c3d4e5f6789012345",
  "status": "not_started",
  "prompt": "Tell us about yourself and your goals.",
  "content": ""
}
```

**Response:**

```json
{
  "success": true,
  "data": {
    "id": "64a1b2c3d4e5f6789012346",
    "title": "Personal Statement",
    "college_application_id": "64a1b2c3d4e5f6789012345",
    "college_name": "Harvard University",
    "status": "not_started",
    "prompt": "Tell us about yourself and your goals.",
    "content": "",
    "ai_score": 0,
    "ai_feedbacks": [],
    "ai_suggestion": "",
    "created_at": "2023-01-01T00:00:00Z",
    "updated_at": "2023-01-01T00:00:00Z"
  },
  "message": "Essay created successfully"
}
```

### Get User Essays

**GET** `/essays`

Returns all essays for the authenticated user with college names included.

### Get Essay by ID

**GET** `/essays/{id}`

Returns a specific essay by ID (only if user owns the associated college application).

### Update Essay

**PUT** `/essays/{id}`

**Request Body:**

```json
{
  "title": "Updated Personal Statement",
  "status": "in_progress",
  "content": "My updated essay content...",
  "ai_score": 8.5,
  "ai_feedbacks": [
    "Great opening paragraph",
    "Consider adding more specific examples"
  ],
  "ai_suggestion": "Try to be more specific about your achievements"
}
```

### Delete Essay

**DELETE** `/essays/{id}`

Soft deletes an essay (sets deleted_at timestamp).

### Get Essays by Status

**GET** `/essays/status?status=in_progress`

Returns essays filtered by status.

### Get Essays by College Application

**GET** `/college-applications/{college_application_id}/essays`

Returns all essays for a specific college application.

### Get Essays with Upcoming Deadlines

**GET** `/essays/deadlines?days=30`

Returns essays whose associated college applications have deadlines within the specified number of days.

### Get Essay Count

**GET** `/essays/count`

Returns the total count of essays for the authenticated user.

## Key Features

1. **Ownership Security**: Essays can only be accessed by users who own the associated college application
2. **College Name Integration**: All essay responses include the college name from the linked college application
3. **AI Integration Ready**: Fields for AI scoring, feedback, and suggestions
4. **Status Tracking**: Comprehensive status system to track essay progress
5. **Soft Delete**: Essays are soft-deleted to maintain data integrity
6. **Aggregation Queries**: Uses MongoDB aggregation pipelines for efficient user-based queries

## Database Relationships

Essays are linked to college applications through the `college_application_id` field. The relationship is enforced at the service layer:

1. When creating an essay, the system verifies the college application exists
2. When accessing essays, the system uses aggregation to join with college applications and filter by user ownership
3. College application names are included in essay responses for better UX

## Error Handling

The system uses consistent error handling:

- **400 Bad Request**: Invalid input data or validation errors
- **404 Not Found**: Essay not found or user doesn't have access
- **500 Internal Server Error**: Database or service errors

## Security

- All essay endpoints require JWT authentication
- User ownership is verified through the college application relationship
- Users can only access essays associated with their own college applications
- Soft delete ensures data integrity while allowing "deletion"

## Future Enhancements

- AI integration for automatic scoring and feedback
- Essay templates and prompt suggestions
- Collaboration features for sharing essays with counselors
- Version history tracking
- Export functionality (PDF, Word, etc.)
