(()=>{

    window.Login = ()=>{
        username = document.getElementById("username").value
        password = SHA512(document.getElementById("password").value)

        var xhr = new XMLHttpRequest();
        xhr.open("POST",  "./login/", true);
        xhr.setRequestHeader('Content-Type', 'application/json');
        xhr.onreadystatechange = function() {
            if (xhr.readyState === 4) {
                window.CallbackLogin(xhr.response)
            }
        }
        xhr.send(JSON.stringify({
            Username: username,
            Password: password
        }));
    }

    window.RegisterUser = ()=>{
        username = document.getElementById("username").value
        password = document.getElementById("password").value
        validatePassword = document.getElementById("validate-password").value
        if(password === validatePassword){
            var xhr = new XMLHttpRequest();
            xhr.open("POST",  "./register/", true);
            xhr.setRequestHeader('Content-Type', 'application/json');
            xhr.onreadystatechange = function() {
                if (xhr.readyState === 4) {
                    window.CallbackRegister(xhr.response)
                }
            }
            xhr.send(JSON.stringify({
                Username: username,
                Password: SHA512(password)
            }));
        }
    }

    window.CallbackLogin = (response) => {
        response = JSON.parse(response)
        if(response.Response == "success"){
            location.href = "./posts"

        } else {
            createMessage(response.Response);
            for(error in response.Error){
                createMessage(response.Errors[error])
            }
        }
    }

    window.CallbackRegister = (response) => {
        response = JSON.parse(response)
        if(response.Response == "success"){
            location.href = "./login"
        } else {
            createMessage(response.Response);
        }
    }
})()