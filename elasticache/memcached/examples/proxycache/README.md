# Proxy Cache

This is a simple example that shows how to use our Memcached service acorn. It proxies requests to a given URL and caches the response using Memcached.

## Example usage

Open `$ACORN_URL/?url=https://files.catbox.moe/wwsyqi.jpg` to see a 4K wallpaper. The first request takes around ~3.5 seconds (~200ms more than the source) for me while the second takes ~1.3 seconds. 

## Running

`acorn run .` in this directory.
