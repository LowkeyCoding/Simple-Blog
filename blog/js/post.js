window.deletePost = (ID) => {
    var answer = window.confirm("Are you sure you want to delete this post?");
    if (answer) {
        var xhr = new XMLHttpRequest();
        xhr.open("POST",  "../deletePost/" + ID, true);
        xhr.setRequestHeader('Content-Type', 'application/json');
        xhr.onreadystatechange = function() {
            if (xhr.readyState === 4) {
                location.href = xhr.response
            }
        }
        xhr.send();
    }
}