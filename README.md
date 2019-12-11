# split-go-serializer

A GO module which fetches split definitions and segments from Split.io and serializes them into a set of strings that the GO SDK can consume.

## Setting Up Dev Environment

1. Setup Go on your local machine by following [these docs](https://golang.org/doc/install#install)
1. Clone this repo
    ```
    $ git clone https://github.com/godaddy/split-go-serializer.git
    ```
1. Install Dependencies in the project directory
    ```
    $ go get ./...
    $ go get golang.org/x/lint/golint
    $ go get github.com/axw/gocov/gocov
    $ GO111MODULE=off go get gopkg.in/matm/v1/gocov-html
    $ go mod tidy
    ```

## Testing

Use this script to run linting, vetting, unit tests, and coverage check:
```
$ ./script/test.sh
```
Add the `-i` option to do all of the above and run the integration tests:
```
$ ./script/test.sh -i
```

After running either of the above scripts, run this script to generate coverage HTML file:
```
$ ./script/coverage-html.sh cover.out > coverage.html
```
This HTML file is useful because it highlights exact lines of code that aren't covered by tests.

## License

[MIT](LICENSE)
