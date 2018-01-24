# configure

Easily fill in a configuration struct with either flag arguments or a JSON blob argument.


## Usage

Given a simple program that uses two arguments, `district_id` and `collection`:

```go
var config struct {
	DistrictID string `config:"district_id,required"`
	Collection string `config:"collection"`
}

func main() {
	if err := configure.Configure(&config); err != nil {
		log.Fatalf("err: %#v", err)
	}
	log.Printf("config: %#v", config)

	// go use arguments to do something
}
```

It can be invoked two styles:

```bash
./example -h
Usage of ./example:
  -collection string
      generated field
  -district_id string
      generated field

./example -district_id="abc123"
> config: {DistrictID:abc123 Collection:}

./example '{"district_id":"abc123","collection":"schools"}'
> config: {DistrictID:abc123 Collection:schools}

./example -district_id="abc123" -collection="schools"
> config: {DistrictID:abc123 Collection:schools}

# fails when not provided with the required district_id argument
./example -collection=schools
> err: Missing required fields: [district_id]

./example '{"collection":"schools"}'
> err: Missing required fields: [district_id]
```

### Defining defaults

You can also define defaults by passing a pre-populated struct:

```go
func main() {
	config := struct {
		DistrictID string `config:"district_id,required"`
		Collection string `config:"collection"`
	}{
		Collection: "default-collection",
	}

	if err := configure.Configure(&config); err != nil {
		log.Fatalf("err: %#v", err)
	}
	log.Printf("config: %#v\n", config)

	// go use arguments to do something
}

```
