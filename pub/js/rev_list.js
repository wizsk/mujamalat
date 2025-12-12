const dialog = document.getElementById("dialog");

document.querySelectorAll(".row").forEach((r) => {
    r.addEventListener('click', () => {
        showModal(r);
    })
})

const dw = document.getElementById("dw");
const df = document.getElementById("df");
// const dp = document.getElementById("dp");
const dh = document.getElementById("dh");
const dfres = document.getElementById("dfreset"); // d future reset

function showModal(r) {
    const word = r.querySelector("[data-w]");
    const future = r.querySelector("[data-f]");
    const past = r.querySelector("[data-p]");
    const dontShow = r.querySelector("[data-h]");
    // console.log(word, future, past, dontShow);

    dw.innerText = word.dataset.w;
    df.innerText = future.dataset.f ? future.dataset.f : "Not set";

    let hidden = dontShow.dataset.h == "true";
    dh.innerText = hidden ? "Show" : "Hide";

    dh.onclick = () => {
        hidden = !hidden;
        dh.innerText = hidden ? "Show" : "Hide";
        dontShow.dataset.h = hidden ? "true" : "false";
        dontShow.innerText = hidden ? "Hidden" : "Shown";
        console.log(dontShow)
    }

    dfres.onclick = () => {
        future.dataset.f = "";
        future.dataset.fu = "0";
        future.innerText = ""
        past.dataset.p = "";
        past.dataset.pu = "0";
        past.innerText = ""
        df.innerText = "Not set";
    }

    dialog.showModal();
}
