# Logparsetest
This app will scan a provided directory (argument 1 to the app), looking for log files.

It will scan them, looking for a predefined regex to determine the accountID.

Within those lines that have accountIDs, it will collect the date field and parse it into a unix epoch.

To calculate the session spans, each page hit's epochs are aligned and tested.  If there is a delta of 600 or more (10 minutes), it will indicate a newsession.  It will also determine the max and min length of sessions.

# Running the app

### Install go tools
https://golang.org/dl/

### Collect dependencies (if necessary)
```
go get
```

## Run it as a script (compile at runtime)
```
go run parser.go logs
```
output: 
```
go run parser.go logs/
Total unique users: 33
Top users:
id		# pages	# sess	longest	shortest
489f3e87	14555	1	860	860
71f28176	8835	1	860	860
95c2fa37	4732	1	857	857
eaefd399	4312	1	857	857
43a81873	3926	3	409	1
```


## Compile it as a distributable app
```
go build
```

output:
```
go build
./logparsetest logs
Total unique users: 33
Top users:
id		# pages	# sess	longest	shortest
489f3e87	14555	1	860	860
71f28176	8835	1	860	860
95c2fa37	4732	1	857	857
eaefd399	4312	1	857	857
43a81873	3926	3	409	1
```


