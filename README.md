# Go To Meet

A MacOS menu bar to go the next meeting in your Google Calendar.

## Configuration

To use this application, you'll need to set up a Google Cloud Project and enable the Google Calendar API:

- Go to the Google Cloud Console
- Create a new project
- Enable the Google Calendar API
- Create OAuth 2.0 credentials
- Add `https://www.googleapis.com/auth/calendar.readonly` to scopes
- Download the credentials and save them as credentials.json

The credentials.json file should look like this:

```json
{
  "client_id": "your-client-id.apps.googleusercontent.com",
  "client_secret": "your-client-secret"
}
```
