#!/bin/bash
# Linux versiobn of build.ps1
# Builds the frontend with pnpm then builds and runs the backend server with go
cd frontend
pnpm install
pnpm run build
cd ..
go build -o GoogleCalendarTimetable ./src/main.go
./GoogleCalendarTimetable
