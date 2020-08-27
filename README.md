Elite Dangerous Market Data Translation
=======================================

A key part of [GoMenacing](https://github.com/kfsone/gomenacing) will be obtaining and importing
data from community sites. When developing [TradeDangerous](https://bitbucket.org/kfsone/tradedangerous)
I ended up with the application investing a lot of time and effort into parsing, importing and
loading data.

The idea is for it to, instead, use Translators to convert foreign formats into a protobuf-based
binary format, and ultimately storing binary encoded protobuf messages into its datastores.

For now this is happening in a side project, focused on eddb, but ultimately a chunk of this code
will migrate over to GoMenacing as a "translators" package/module.

Protobufs is cross-platform, cross-medium, so it lends itself well to building distributed services.
