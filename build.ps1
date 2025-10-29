# Builds the frontend with pnpm then builds and runs the backend server with go
Set-Location frontend
pnpm install
pnpm run build
Set-Location ..
go build -o GoogleCalendarTimetable.exe ./src/main.go
./GoogleCalendarTimetable.exe