# gcfg
Fork of Péter Surányi's INI-style configuration file parser in Go.

Original project page: [https://code.google.com/p/gcfg](https://code.google.com/p/gcfg)

[![Build Status (Linux)](https://travis-ci.org/baobabus/gcfg.svg?branch=master)](https://travis-ci.org/baobabus/gcfg)

# Additional Features

## Custom type parsers

Custom parsers can be registered with gcfg runtime. This is usually needed for "imported" types that do not implement text unmarshalling.

The below example enables parsing of time.Duration and url.URL types:

```go
package main

import (
	"github.com/baobabus/gcfg"
	"net/url"
	"time"
)

func init() {
	var d time.Duration
	gcfg.RegisterTypeParser(reflect.TypeOf(d), func(blank bool, val string) (interface{}, error) {
		if blank {
			return nil, nil
		}
		return time.ParseDuration(val)
	})
	gcfg.RegisterTypeParser(reflect.TypeOf(url.URL{}), func(blank bool, val string) (interface{}, error) {
		if blank {
			return nil, nil
		}
		return url.Parse(val)
	})
}
```

```
[remote]
url = http://foo.com/bar/
timeout = 1m
```

## Optional configuration sections

Configuration sections can be pointers to structs. The section structure will only be allocated and the reference updated if the input explicitly specifies the section.

In the below example, if Secondary is initially nil, it will remain nil unless [secondary] is explicitly specified in the input .ini file.

```go
type Gateway struct {
	Url     *url.URL
	Timeout time.Duration
}

type Config struct {
	Primary   Gateway
	Secondary *Gateway
}
```

## Constraints

Ordered types can have bounds specified:

```go
type Gateway struct {
	Url     *url.URL
	Timeout time.Duration `min:"5s" max:"15m"`
}
```

Strings can have their length constrained:

```go
type Config struct {
	Message string `minlen:"1" maxlen:"400"`
}
```

## Protected fields

Configuration section fields can be locked down. This prevents the field from being set from .ini file.

The following definition:

```go
type Config struct {
	Foo string `gcfg:"-"`
}
```

will produce an error when reading config:

```
[config]
foo = blah
```

## Basic support for writing out the configuration

With certain limitations, runtime structures can be written out in .INI format.

```go
	gcfg.Write(&myConfig, os.Stdout)
```
