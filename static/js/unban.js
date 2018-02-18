(function () {
    document.addEventListener('DOMContentLoaded', init, false);
    function init() {
        for (let e of document.querySelectorAll('.input')) {
            e.addEventListener('click', submitted);
        }
    }
    function submitted(event) {
        let root = event.srcElement.parentElement.parentElement;
        let ipAddr = document.querySelectorAll('.row')[root.rowIndex - 1].querySelectorAll('td')[4].innerHTML;
        event.preventDefault();
        let request = new XMLHttpRequest();
        request.onreadystatechange = () => {
            if (request.readyState === 4) {
                if (request.status === 200) {
                    window.alert(ipAddr + " has been unbanned.");
                    location.reload();
                }
                else {
                    window.alert("Uh oh! Something didnt go right :/");
                }
            }
            else {
                return;
            }
        };
        request.open('DELETE', 'unban?ip=' + ipAddr, true);
        request.send();
    }
}());
//# sourceMappingURL=unban.js.map