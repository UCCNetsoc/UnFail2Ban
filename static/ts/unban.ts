(function () {
    document.addEventListener('DOMContentLoaded', init, false);
   
    var rows: NodeListOf<Element>;
    var tableDiv: HTMLElement;
   
    function init() {
        rows = document.querySelectorAll('.row');        
        for(let e of document.querySelectorAll('.input')) {
            e.addEventListener('click', submitted)
        }
        tableDiv = document.getElementById("table");
    }
    
    function submitted(event: MouseEvent) {
        let root = event.srcElement.parentElement.parentElement as HTMLTableRowElement
        let ipAddr = rows[root.rowIndex-1].querySelectorAll('td')[4].innerHTML;
        
        event.preventDefault()

        let request = new XMLHttpRequest();
        request.onreadystatechange = () => {
            if (request.readyState === 4 && request.status === 200) {
                tableDiv.innerHTML = request.responseText;
                init();
                window.alert(ipAddr + " has been unbanned.");
                return
            }
            window.alert("Uh oh! Something didnt go right :/");
        };
        request.open('GET', 'unban?ip=' + ipAddr, true);
        request.send();
    }
}());