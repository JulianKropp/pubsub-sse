const serverUrl = 'http://localhost:8080/event'; // URL to your EventSource endpoint
let chatService, user, publicTopicId = 'public_chat';

function joinChat() {

}

function sendMessage() {

}

function updateChat(message) {
    const chatBox = document.getElementById('chat');
    chatBox.innerHTML += `<div>${message}</div>`;
    chatBox.scrollTop = chatBox.scrollHeight;
}

window.addEventListener('beforeunload', function () {
    if (chatService && chatService.evtSource) {
        chatService.evtSource.close();
    }
});
