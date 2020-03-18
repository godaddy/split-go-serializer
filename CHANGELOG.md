# Change Log

All notable changes to this project will be documented in this file.

<a name="v3.0.0"></a>
## [v3.0.0](github.com/godaddy/split-go-serializer/compare/v2.0.0...v3.0.0) (2020-03-18)

### Chore

* **ci:** use circleci instead of travis ([#16](github.com/godaddy/split-go-serializer/issues/16))
* **release:** 3.0.0

### Feat

* **module:** major change ([#22](github.com/godaddy/split-go-serializer/issues/22))

### Fix

* **Fetcher:** update method to have a string list parameter ([#23](github.com/godaddy/split-go-serializer/issues/23))
* **circleci:** fix test command ([#17](github.com/godaddy/split-go-serializer/issues/17))


<a name="v2.0.0"></a>
## [v2.0.0](github.com/godaddy/split-go-serializer/compare/v1.0.2...v2.0.0) (2020-01-16)

### Chore

* **mod:** update to v2 ([#15](github.com/godaddy/split-go-serializer/issues/15))
* **release:** 2.0.0
* **release:** 2.0.0

### Refactor

* **api:** change splits and segments to maps with keys that match data loader expectations ([#12](github.com/godaddy/split-go-serializer/issues/12))
* **serializer:** handle empty cache and stringify each split and segment ([#13](github.com/godaddy/split-go-serializer/issues/13))

### BREAKING CHANGE


adds an empty object to the script tag when the cache is empty and returns individual split and segment values as strings


<a name="v1.0.2"></a>
## [v1.0.2](github.com/godaddy/split-go-serializer/compare/v1.0.1...v1.0.2) (2020-01-02)

### Chore

* **release:** 1.0.2

### Refactor

* **serializer:** handle empty cache similar to node-serializer ([#11](github.com/godaddy/split-go-serializer/issues/11))


<a name="v1.0.1"></a>
## [v1.0.1](github.com/godaddy/split-go-serializer/compare/v1.0.0...v1.0.1) (2019-12-30)

### Chore

* **release:** 1.0.1

### Fix

* **api:** create decoder that sets tagName to json ([#10](github.com/godaddy/split-go-serializer/issues/10))


<a name="v1.0.0"></a>
## v1.0.0 (2019-12-27)

### Chore

* **release:** 1.0.0

### Feat

* **api:** implement getAllChanges and GetSplits function ([#4](github.com/godaddy/split-go-serializer/issues/4))
* **apiBinding:** implement functions to get segments ([#6](github.com/godaddy/split-go-serializer/issues/6))
* **poller:** add poller to periodically fetch data ([#3](github.com/godaddy/split-go-serializer/issues/3))
* **poller:** implement pollForChanges function ([#7](github.com/godaddy/split-go-serializer/issues/7))
* **readme:** add usage and versioning ([#9](github.com/godaddy/split-go-serializer/issues/9))
* **segment:** add getSegmentNamesInUse function ([#5](github.com/godaddy/split-go-serializer/issues/5))
* **serializer:** implement getSerializedData function ([#8](github.com/godaddy/split-go-serializer/issues/8))
* **split-go-serializer:** setup and outline classes ([#1](github.com/godaddy/split-go-serializer/issues/1))
* **splitio-api-wrapper:** implement httpGet function ([#2](github.com/godaddy/split-go-serializer/issues/2))

