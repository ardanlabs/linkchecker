# LinkChecker

Recurses through web site and reports any broken links. Meant to be used as part of the continuous
integration process.

# Usage

Run and provide a host.

```
./linkchecker -host=www.ardanstudios.com
```
Run against host without valid certificate

```
./linkchecker -host=www.ardanstudios.com -skiptls
```

# Developing

Run make to build and run locally and test against `www.ardanstudios.com`.

```
make run
```

# Docker

To build and upload a Docker image to quay.io use the make command.

```
make docker
```

You can run the `linkchecker` from Docker like so.

```
docker run --rm -e HOST=www.ardanlabs.com quay.io/ardanlabs/linkchecker
```
