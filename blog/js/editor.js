// await the loading of the web assembly module
window["markdown"].ready.then(markdown => {
    const title = document.getElementById("title")
    const description = document.getElementById("description-input")
    const md = document.getElementById("markdown-input")
    const html = document.getElementById("html-output")
    const publish = document.getElementById("publish")
    let hljsTimer

    function update() {
        let source = "#"+title.value+"\n"+md.value
        let _html = markdown.parse(source, {
            parseFlags: markdown.ParseFlags.DEFAULT | markdown.ParseFlags.NO_HTML,
        })
        html.innerHTML = _html
        window.updateCodeHighlight()
        md.style.height = (html.clientHeight - 32) + "px"
    }

    window.updateCodeHighlight = () => {
        clearTimeout(hljsTimer)
        if (typeof hljs == "undefined") {
            hljsTimer = setTimeout(updateCodeSyntaxHighlighting, 500)
            return
        }

        document.querySelectorAll('pre code[class^="language-"]').forEach(block => {
            hljs.highlightBlock(block)
        })
    }

    window.publishPost = () => {
        _html = markdown.parse(document.getElementById("markdown-input").value, {
            parseFlags: markdown.ParseFlags.DEFAULT | markdown.ParseFlags.NO_HTML,
        })
        var xhr = new XMLHttpRequest();
        xhr.open("POST",  "../newPost", true);
        xhr.setRequestHeader('Content-Type', 'application/json');
        xhr.onreadystatechange = function() {
            if (xhr.readyState === 4) {
                response = JSON.parse(xhr.response)
                location.href = response.Response
            }
        }
        xhr.send(JSON.stringify({
            Title: title.value,
            Description: description.value,
            Content: md.value,
            HTML: _html,
        }));
    }

    window.updatePost = (id) => {
        _html = markdown.parse(document.getElementById("markdown-input").value, {
            parseFlags: markdown.ParseFlags.DEFAULT | markdown.ParseFlags.NO_HTML,
        })
        html.innerHTML = _html
        window.updateCodeHighlight()
        var xhr = new XMLHttpRequest();
        xhr.open("POST",  "../updatePost/"+id, true);
        xhr.setRequestHeader('Content-Type', 'application/json');
        xhr.onreadystatechange = function() {
            if (xhr.readyState === 4) {
                response = JSON.parse(xhr.response)
                location.href = response.Response
            }
        }
        xhr.send(JSON.stringify({
            ID: id,
            Title: title.value,
            Description: description.value,
            Content: md.value,
            HTML: html.innerHTML,
        }));
    }

    md.addEventListener("input", update)
    title.addEventListener("input", update)
    window.addEventListener('resize', ()=>{
        md.style.height = (html.clientHeight - 32) + "px"
    })
    if(title.value != ""){
        update()
        md.style.height = (html.clientHeight - 32) + "px"
    }

    // load the markdown module to make it playable in the browser console
    console.log("markdown module API:", markdown)
})