# split-go-serializer

A Go module which fetches split definitions and segments from Split.io and serializes them into a set of strings that the GO SDK can consume.

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
## Usage

Use this Go module in your server-side Go environment. The serializer exposes:
1. a `Poller` that periodically requests raw experiment configuration data from Split.io. Requests happen in the background and the poller caches the latest data in local memory.

### Instantiation

Create an instance of `Poller` by importing the `poller` package of this module and calling the `NewPoller` function with some parameters :

```go
import (
    "github.com/godaddy/split-go-serializer/poller"
)

poller := poller.NewPoller("YOUR_API_KEY", 600, false, nil)
```

The following option properties are available to the `Poller`:

| Property                      | Description |
|-------------------------------|-------------|
| splitioApiKey | The Split.io SDK key for the environment your app is running in. (required) |
| pollingRateSeconds | The interval at which to poll Split.io. Defaults to 300 (5 minutes). |
| serializeSegments | Whether or not to fetch segment configuration data. Defaults to false.|

#### Serializing segments

Segments are pre-defined groups of customers that features can be targeted to. More info [here](https://help.split.io/hc/en-us/articles/360020407512-Create-a-segment).

**Note:** Requesting serialized segments will increase the size of your response. Segments can be very large if they include all company employees, for example.

### Methods

#### Start

Make an initial request for changes and start polling for raw configuration data
every `pollingRateSeconds`:

```go
poller.Start()
```

#### Stop
To stop the poller:

```go
poller.Stop()
```

The poller sends an error message to `poller.Error` channel when getting errors from the Split.io API.

#### GetSerializedData

`GetSerializedData` will read the latest data from the cache and return a script
that adds serialized data to the `window.__splitCachePreload` object. The
serialized data will be used to determine cohort allocations.

`GetSerializedData` accepts the following arguments:

| Property                      | Description |
|-------------------------------|-------------|
| splitNames | Array of strings that, if non-empty, filters the `splitsData` |

```go
serializedDataScript := poller.GetSerializedData([]string{})
fmt.Println(serializedDataScript)

//<script>
//  window.__splitCachePreload = {
//    splitsData: {
//      "split-1-name":"{\"name\":\"split-1-name\",\"status\":\"bar\"}",
//      "split-2-name":"{\"name\":\"split-2-name\",\"status\":\"baz\"}"
//    },
//    since: 1,
//    segmentsData: {
//       "test-segment":"{\"name\":\"test-segment\",\"added\":[\"foo\",\"bar\"],\"removed\":null,\"since\":20,\"till\":20}"
//    },
//    usingSegmentsCount: 2
//  }
//</script>

serializedDataScript = poller.GetSerializedData([]string{"split-1-name"})
fmt.Println(serializedDataScript)

//<script>
//  window.__splitCachePreload = {
//    splitsData: {
//      "split-1-name":"{\"name\":\"split-1-name\",\"status\":\"bar\"}"
//    },
//    since: 1,
//    segmentsData: {
//       "test-segment":"{\"name\":\"test-segment\",\"added\":[\"foo\",\"bar\"],\"removed\":null,\"since\":20,\"till\":20}"
//    },
//    usingSegmentsCount: 2
//  }
//</script>
```

## Testing

Use this script to run linting, vetting, unit tests, and coverage check:
```
$ ./scripts/test.sh
```

After running the above script, run this script to generate coverage HTML file:
```
$ ./scripts/coverage-html.sh cover.out > coverage.html
```
This HTML file is useful because it highlights exact lines of code that aren't covered by tests.

## Module Versioning

We utilize [`git-chglog`](https://github.com/git-chglog/git-chglog) to maintain our CHANGELOG.
Please follow these steps on the `master` branch of this repository to update the Go module version with changes made since the prior version.

1. Ensure `git-chglog` is installed correctly: [docs](https://github.com/git-chglog/git-chglog#installation)
1. Fetch all version tags
    ```
    $ git fetch --tags
    ```
1. Tag with the new version based on [`semvar conventions`](https://semver.org/)
    ```
    $ git tag v{MAJOR}.{MINOR}.{PATCH}
    ```
1. Run `git-chglog` to generate new `CHANGELOG.md` file
    ```
    $ git-chglog -o CHANGELOG.md
    ```
1. Stage and commit changes to the `CHANGELOG`
1. Re-tag to accomodate commit created in the previous step
    ```
    $ git tag -d v{MAJOR}.{MINOR}.{PATCH}
    $ git tag v{MAJOR}.{MINOR}.{PATCH}
    ```
1. Push to origin tag and origin master
    ```
    $ git push origin v{MAJOR}.{MINOR}.{PATCH}
    $ git push origin master
    ```

## License

[MIT](LICENSE)
