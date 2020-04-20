# spanner-opencensus-example

This shows an example to enable opencensus metrics in cloud spanner client library. 

## Run

You need to set project, instance, and database for a Cloud Spanner database. 

```bash
export PROJECT_ID="YOUR_PROJECT_ID"
export INSTANCE_ID="YOUR_INSTANCE_ID"
export DATABASE_ID="YOUR_DATABASE_ID"
```

Then, you can run the api server:

```bash
$ go run webapp_enable_oc.go // The default port is 8080.

// Or, specify a port.

$ go run webapp_enable_oc.go -port=8080
```
