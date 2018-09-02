(function() {
  document.addEventListener('DOMContentLoaded', init, false);
  
  var logDiv: HTMLDivElement
  var date: string
  
  function init(){
    logDiv = document.getElementById("log") as HTMLDivElement
    
    poll()
    setInterval(poll, 5000)
  }
  
  function poll(){
    const request = new XMLHttpRequest()
    request.onreadystatechange = () => {
      if(request.readyState !== 4){
        return
      }

      if(request.status === 200){  
        if(request.responseText != "") {
          logDiv.innerHTML += request.responseText
          const splitText = request.responseText.split("\n")
          date = splitText[splitText.length-2].substr(0,23)
          logDiv.scrollTop = logDiv.scrollHeight
        }
      }
    }
    request.open('GET', 'poll?date="'+date+'"', true)
    request.send()
  }
}())