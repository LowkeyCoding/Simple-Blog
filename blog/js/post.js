window.deletePost = (ID) => {
    var answer = window.confirm("Are you sure you want to delete this post?");
    if (answer) {
        var xhr = new XMLHttpRequest();
        xhr.open("POST",  "../deletePost/" + ID, true);
        xhr.setRequestHeader('Content-Type', 'application/json');
        xhr.onreadystatechange = function() {
            if (xhr.readyState === 4) {
                response = JSON.parse(xhr.response)
                for(error in response.Errors){
                    createMessage(response.Errors[error])
                }
                console.log(response)
                if (response.Response != ""){
                    location.href = response.Response
                }
            }
        }
        xhr.send();
    }
}