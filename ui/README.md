# quorum-reporting-ui

Basic UI for quorum reporting engine

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

Currently, (statik)[https://github.com/rakyll/statik] is used for this; refer to their documentation 
on how to install it.

To generate the production resource, run the following from within the UI folder:
- `npm install`
- `npm run-script build`
- `statik -src=./build -f`

This will have updated the assets file under `statik/statik.go` and the new resources
will be compiled into the application.