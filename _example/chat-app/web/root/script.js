function joinChatRoom() {
    var username = document.getElementById('username').value;
    var chatroom_id = document.getElementById('chatroom_id').value;

    var xhr = new XMLHttpRequest();
    xhr.open("POST", "/join", true);
    xhr.setRequestHeader("Content-Type", "application/json");
    xhr.onreadystatechange = function () {
        if (xhr.readyState === 4) {
            if (xhr.status === 200) {
                var response = JSON.parse(xhr.responseText);
                window.location.href = `/chatroom?client_id=${response.client_id}&chatroom_id=${chatroom_id}`;
            } else {
                document.getElementById('error-message').innerText = "Room wasn't found.";
            }
        }
    };
    var data = JSON.stringify({ "name": username, "chatroom_id": chatroom_id });
    xhr.send(data);
}
