![Build passing](https://github.com/antpas14/fantalegheGO/actions/workflows/master.yaml/badge.svg)
[![codecov](https://codecov.io/gh/antpas14/fantalegheGO/graph/badge.svg?token=M2129SSBZJ)](https://codecov.io/gh/antpas14/fantalegheGO)

# FantalegheGO

This is a golang implementation of [fantalegheEV project](https://github.com/antpas14/fantalegheEV-api)

To summarise this project is a web application that permits to recalculate a fantasy league rank in a *fair* way.


This application backend uses [echo](https://github.com/labstack/echo), a minimalistic web server and [goQuery](github.com/PuerkitoBio/goquery) to parse HTTP returned from <a href="https://github.com/antpas14/webFetcher">webFetcher</a>, a simple app that returns the HTML of the requested page after some javascript rendering is executed.

A docker compose is provided which also utilises a basic UI that can be found <a href="https://github.com/antpas14/fantalegheEV-ui">here</a>

This application analyzes football fantasy league hosted on <a href="http://leghe.fantacalcio.it">leghe.fantacalcio.it</a>. I have no relationship with them.

### License

This work is distributed under MIT license.
