mite
====

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
