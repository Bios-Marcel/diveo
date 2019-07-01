# Diveo

This is a tiny video format I made because I was bored. The goal here was to
try and make a rather small video format for videos where not many pixels
between each frame differ. While this might be useless, it was certainly fun
and I learned a thing or two.

## Specification

The specification can be found at [SPECv1.md](spec/SPECv1.md).

## Example implementation

An example recorder that runs on linux using X11 can be found in the
[subdirectory demo](/demo). It requires `xrectsel` to be present on the host
that executes it.

Currently there is no player for the format.
