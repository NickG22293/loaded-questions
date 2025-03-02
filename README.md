# loaded-questions
Web app backed by Go, UI in React that runs a simple session-based questions + answer game

# Setup 
.devcontainer folder contains definition for containerized development env. You need the Devcontainers extension in VSCode installed for this. 

```
cd .devcontainer
docker build -t dev-go-node . 
``` 
Reopen in a container via VSCode. 

# Backend 
## Build
```
cd backend
go build -o main .
```

## Test
```
cd backend
go test ./...
```

## Run
```
cd backend
./main
```

## Docker
### Build
```
cd backend
docker build -t loaded-questions-backend .
```

### Run
```
docker run -p 8080:8080 loaded-questions-backend
```

# Frontend
## Build
```
cd frontend
npm install
npm run build
```

## Test
```
cd frontend
npm install
npm test
```

## Run
```
cd frontend
npm install
npm start
```