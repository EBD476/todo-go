# Google Drive Setup Guide

This guide will help you set up Google Drive integration for the Todo CLI application.

## Prerequisites

- A Google account
- Go 1.21 or later installed
- Internet connection

## Step 1: Create a Google Cloud Project

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Click "Select a project" or "New Project"
3. Click "New Project"
4. Enter a project name (e.g., "todo-cli-drive")
5. Click "Create"

## Step 2: Enable Google Drive API

1. In the Google Cloud Console, make sure your new project is selected
2. Go to "APIs & Services" > "Library"
3. Search for "Google Drive API"
4. Click on "Google Drive API"
5. Click "Enable"

## Step 3: Create Credentials

1. Go to "APIs & Services" > "Credentials"
2. Click "Create Credentials" > "OAuth client ID"
3. If prompted, configure the OAuth consent screen:
   - Choose "External" user type
   - Fill in the required fields (App name, User support email, Developer contact)
   - Add your email to test users
   - Save and continue through the steps
4. For Application type, choose "Desktop application"
5. Give it a name (e.g., "Todo CLI")
6. Click "Create"
7. Download the JSON file and rename it to `credentials.json`
8. Place `credentials.json` in the same directory as your `todo` executable

## Step 4: Install Dependencies

Run the following command to install the required Go packages:

```bash
go mod tidy
```

## Step 5: Build and Test

1. Build the application:
   ```bash
   go build -o todo main.go
   ```

2. Test the Google Drive integration:
   ```bash
   # Upload your todos to Google Drive
   ./todo upload
   
   # Download todos from Google Drive
   ./todo download
   ```

## First Time Setup

When you run `./todo upload` for the first time:

1. The application will open your browser
2. Sign in to your Google account
3. Grant permissions to the application
4. Copy the authorization code from the browser
5. Paste it into the terminal
6. The application will save a `token.json` file for future use

## Commands

- `todo upload` or `todo up` - Upload todos to Google Drive
- `todo download` or `todo down` - Download todos from Google Drive

## File Management

- The app will create/update a file called `todos-backup.json` in your Google Drive
- Local todos are still stored in `todos.json`
- The Google Drive file serves as a cloud backup

## Troubleshooting

### "credentials not found" error
- Make sure `credentials.json` is in the same directory as the `todo` executable
- Verify the file name is exactly `credentials.json` (case-sensitive)

### "Unable to read authorization code" error
- Make sure you copy the entire authorization code from the browser
- The code should be a long string of characters

### "Failed to get Google Drive service" error
- Check your internet connection
- Verify the Google Drive API is enabled in your Google Cloud project
- Make sure the credentials file is valid JSON

### Token expired
- Delete the `token.json` file and run `./todo upload` again
- This will prompt you to re-authenticate

## Security Notes

- Keep your `credentials.json` file secure and don't share it
- The `token.json` file contains your access token - keep it secure too
- Both files should be added to `.gitignore` if you're using version control

## Features

- **Automatic file detection**: The app will update existing files or create new ones
- **OAuth 2.0 authentication**: Secure authentication with Google
- **Token caching**: No need to re-authenticate every time
- **Error handling**: Clear error messages and troubleshooting guidance
- **Beautiful CLI**: Consistent with the rest of the todo app's design

## Support

If you encounter issues:

1. Check this guide first
2. Verify all steps were completed correctly
3. Check the Google Cloud Console for any API quota issues
4. Ensure your Google account has sufficient storage space
