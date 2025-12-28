# Builds the frontend with pnpm then builds and runs the backend server with go
Set-Location frontend
pnpm install
pnpm run build
Set-Location ..
go build -o GoogleCalendarTimetable.exe ./backend/main.go
./GoogleCalendarTimetable.exe