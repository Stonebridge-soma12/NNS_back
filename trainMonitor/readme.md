## Usage
```go
func hanlder(r *http.Response, w http.ResponseWriter) {
	edr EpochDbRepository {
		DB: db,
	}
    
	err := trainMonitor.PostEpochHandler(r, edr)
	if err != nil {
		error handling ...
	}
}
```