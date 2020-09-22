Basic UI for Quorum Reporting engine

## Getting Started

Install dependencies

`npm install`  

Start development server

`npm start`  

Open UI

`http://localhost:3000`  

## Generating production assets

In order to have the assets be part of the application binary, they
must be packaged up and included in the build process.

Currently, (statik)[https://github.com/rakyll/statik] is used for this purpose. Install it with:
```
go get github.com/rakyll/statik
```

To generate the production resources, run the following from within the UI folder:
```
npm install
npm run-script build
statik -src=./build -f
```

This will update the assets file under `statik/statik.go` and the new resources
will be compiled next time you build the reporting application.
