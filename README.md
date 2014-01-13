# mite

Minimal Templating language inspired by slim and jade, implemented in Go.

Example:

    html
        body
            h1 class='intro' Hello, World!
            p ` class='not an attribute' but text due to the back tick
            div
                `	This is a mite template that can also handle multiple lines 
                    thanks to the backtick in an indented "tag"


Output with indentation for display. Tags in real output are collapsed:

    <html>
        <body>
            <h1 class='intro'>
                Hello, World!
            </h1>
            <p>
                class='not an attribute' but text due to the back tick
            </p>
            <div>
                This is a mite template that can also handle multiple lines 
                thanks to the backtick in an indented "tag"
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
- filter blocks of text (example: for markdown)
- function calling
