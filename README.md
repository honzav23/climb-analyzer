# ClimbAnalyzer

A small web app that analyzes climbs on a bike trip from .gpx file.

## Features
1. Basic trip summary - shows the basic information about a trip (distance, elevation), the elevation profile and the visualization of the route on a map.

<img width="1116" height="668" alt="obrazek" src="https://github.com/user-attachments/assets/903492f7-72dd-4c3f-80c9-96487fed1a35" />

<img width="1082" height="638" alt="obrazek" src="https://github.com/user-attachments/assets/7c369216-ace0-4d73-ba65-1556984c91f7" />

2. Climb analysis - finds significant climbs along the way and displays the basic info (length, elevation, average gradient) together with the elevation profile with different colors based on the section gradient.

<img width="1116" height="669" alt="obrazek" src="https://github.com/user-attachments/assets/318dc0f3-2f31-4cf7-981f-0da94e25e583" />

## How to run

### Using Docker
Run `docker compose up` in the project root to run the app. The app is available at `http://localhost:80`.

### Without Docker - backend
1. Install Go 1.25+.
2. Go to `climb-analyzer-be` folder and type `go mod download` to get the dependencies.
3. Run the server with `go run .`. The server runs at `http:localhost:8080`.

### Without Docker - frontend
1. Install NodeJS 22.17.0+.
2. Go to `climb-analyzer-fe` folder and type `npm install to install the dependencies`.
3. Install Angular by `npm install -g @angular/cli`.
4. Run the frontend app with 'ng serve'. The app runs at `http://localhost:4200`.
