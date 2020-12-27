function createMessage(message){
    let messages = document.getElementsByClassName("messages")[0];
    let alert = document.createElement("div");
    alert.className = "alert";

    let button = document.createElement("span");
    button.className = "closebtn";
    button.addEventListener("click", function() {
        this.parentElement.style.display='none';
    });
    button.innerHTML = "&times;";

    alert.innerHTML = message;
    alert.appendChild(button);

    messages.appendChild(alert);
}

