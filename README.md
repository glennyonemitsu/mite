# mite

Minimal Templating language inspired by slim and jade, implemented in Go.

Example:

    html
        body
            h1 class='really big' Hello, World!
            div
                p This is a mite template


Output (indented for display):

    <html>
        <body>
            <h1 class='really big'>
                Hello, World!
            </h1>
            <div>
                <p>
                    This is a mite template
                </p>
            </div>
        </body>
    </html>

## Goals

Mite aims to be shorthand for html/xml style markup.

It might or might not get compiled into `text/template` compatible format. It
most likely will.

## TODO

- class and id literals `.row#nav`
- variables
- flow control, iteration, and assignment
- explicit string support with backticks
- filter blocks of text (example: for markdown)
- function calling
