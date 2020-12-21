(()=>{
    window.Login = ()=>{
        username = document.getElementById("username").value
        password = SHA512(document.getElementById("password").value)

        var xhr = new XMLHttpRequest();
        xhr.open("POST",  "./login/", true);
        xhr.setRequestHeader('Content-Type', 'application/json');
        xhr.onreadystatechange = function() {
            if (xhr.readyState === 4) {
                window.Callback(xhr.response)
            }
        }
        xhr.send(JSON.stringify({
            Username: username,
            Password: password
        }));
    }
    window.Callback = (response) => {
        console.log(response)
        if(response == "success"){
            location.href = "./posts"

        } else if(response == "failed"){
            alert("Invalid password; please try again.")
        }
    }
})()