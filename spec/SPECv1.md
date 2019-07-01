# Specification - Version 1 - Draft

The data consistens of two data blocks, one being the meta data and the other
being the actual image data. The format doesn't support any audio.

The format currently supports up to 4K resolution.

## Meta-Data

| Position | Content | Size (Bits) |
| - | - | - |
| 0 | The versionnumber as an unsigned 8-bit integer | 8 |
| 8 | The amount of frames per second as an unsigned 8-bit integer | 8 |
| 16 | The length of the video in milliseconds as an unsigned 32-bit integer | 32 |
| 48 | The width of the video as an unsigned 12-bit integer | 12 |
| 60 | The height of the video as an unsigned 12-bit integer | 12 |

## Image-Data

The image-data is rather simple. Each pixel contains of a 24-bit block. The
8-bit parts of the block represent the red, green and blue values of each 
pixel. The first image of the video is complete, meaning that information
regarding each pixel is present. The 24-bit pixel-blocks amount to 
`width*height` blocks. All of those blocks are consequent without any data 
inbetween.

After the first frame, each following frame is just a diff between the previous
frame and the current frame. The diffs are pixelwise and a pixel diff is always
40-bit big. The first 16-bit are the pixel-index in the image and the following
24-bit are the pixel-data. Each frame of the video contains at least a diff
for the first and the last pixel of each frame, in order to show that there
is a frame. So no matter if the first and the last pixel of a frame are 
actually different to the previous frame, the diff will be present.

The image data is current compressed using gzip with the best possible 
compression rate.
