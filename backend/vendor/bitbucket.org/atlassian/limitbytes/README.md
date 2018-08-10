# limitbytes - A middleware to enforce limits on the http request body size.

This is a thin wrapper around `http.MaxBytesReader`. This fails the request early
before reading data if content-length is known, otherwise returns a custom
error type `ErrTooLarge` by parsing the error from `http.MaxBytesReader`. This
also sets the response code to 413 (`http.StatusRequestEntityTooLarge`).

An optional callback can be provided to handle failure case (but this only
works if the content-length is available).

## Usage

```golang
// Create a router that uses std lib http.Handler types for middleware.
var router = chi.NewMux()
router.Use(limitbytes.New(16 * 1024), callback) // limit request body to 16kB
```

When content-length is unknown, the request body has to be read to know if the
limit is breached. For example, this could happen while decoding json from the
body. In such cases, the wrapped http handler can test for `ErrTooLarge` error
to handle them.

```golang
 var wrapped = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	_, e := json.NewDecoder(r.Body).Decode(obj)
	switch e.(type) {
	case nil:
		w.WriteHeader(http.StatusOK)
	case limitbytes.ErrTooLarge:
		w.WriteHeader(http.StatusRequestEntityTooLarge)
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
})
```

## Contributors

Pull requests, issues and comments welcome. For pull requests:

* Add tests for new features and bug fixes
* Follow the existing style
* Separate unrelated changes into multiple pull requests

See the existing issues for things to start contributing.

For bigger changes, make sure you start a discussion first by creating
an issue and explaining the intended change.

Atlassian requires contributors to sign a Contributor License Agreement,
known as a CLA. This serves as a record stating that the contributor is
entitled to contribute the code/documentation/translation to the project
and is willing to have it used in distributions and derivative works
(or is willing to transfer ownership).

Prior to accepting your contributions we ask that you please follow the appropriate
link below to digitally sign the CLA. The Corporate CLA is for those who are
contributing as a member of an organization and the individual CLA is for
those contributing as an individual.

* [CLA for corporate contributors](https://na2.docusign.net/Member/PowerFormSigning.aspx?PowerFormId=e1c17c66-ca4d-4aab-a953-2c231af4a20b)
* [CLA for individuals](https://na2.docusign.net/Member/PowerFormSigning.aspx?PowerFormId=3f94fbdc-2fbe-46ac-b14c-5d152700ae5d)

## License

Copyright (c) 2017 Atlassian and others.
Apache 2.0 licensed, see [LICENSE.txt](LICENSE.txt) file.
