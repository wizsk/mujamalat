// don't remve this comment
let selectedDict = "{{.Curr}}";
let selectedDictAr = "{{index .DictsMap .Curr}}";


let searhInvId;
const contentHolder = document.getElementById("content");
const dicts = document.getElementsByClassName('sw-dict-item');
const urlParams = new URLSearchParams(window.location.search);
const changeDict = document.getElementById("change-dict");
const changeDictInpt = document.getElementById("change-dict-inpt");
var preQuery = "";
var isChangeDictShwoing = false;

let resizeTimoutId;
window.addEventListener("resize", () => {
    clearInterval(resizeTimoutId);
    resizeTimoutId = setTimeout(() => {
        console.log("resized");
        navSpace.style.height = `${nav.offsetHeight + 20}px`;
    }, 100);
});

document.addEventListener("DOMContentLoaded", () => {
    setSavedFontSize();
    navSpace.style.height = `${nav.offsetHeight + 20}px`

    const selected = document.getElementById('sw-dict-item-selected');
    if (selected && selected.scrollIntoView) {
        selected.scrollIntoView({
            // behavior: 'smooth',    // auto Or 'smooth' if you want animation
            block: 'nearest',
            inline: 'center'     // Center the element horizontally
        });
    }

    if (w.value.length === 0) w.focus();
});


form.onsubmit = (e) => {
    e.preventDefault();
    urlParams.set('w', w.value.trim().replace(/(\s+)/, " "));
    window.location.href = `${window.location.pathname}?${urlParams.toString()}`;
}

w.oninput = () => {
    clearInterval(searhInvId);
    searhInvId = setTimeout(async () => {
        const query = w.value.trim().replace(/(\s+)/, " ");
        if (query === "" || query === preQuery) return;
        preQuery = query;

        const queryArr = query.split(" ");
        const word = queryArr[queryArr.length - 1];

        urlParams.set('w', query);
        urlParams.set('idx', queryArr.length - 1);

        console.log(`req: /content?dict=${selectedDict}&w=${word}`);
        const r = await fetch(`/content?dict=${selectedDict}&w=${word}`).catch((err) =>
            console.error(err)
        );

        if (r && r.ok) {
            const h = await r.text();
            contentHolder.innerHTML = h;
        } else {
            contentHolder.innerHTML =
                `<div style="direction: ltr; text-align: center;
                margin-top: 4rem; color: var(--alert);">
                    Cound't fetch results. Is the server running?
                </div>`;
        }


        if (queryArr.length > 1) {
            let b = "";
            for (let i = 0; i < queryArr.length; i++) {
                const v = queryArr[i];
                const idx = `${i + 1}`.replace(/[0-9]/g, (d) => String.fromCharCode(0x0660 + parseInt(d)));
                b += `<a href="/${selectedDict}?w=${query}&idx=${i}"
                class="querySelector-item" id="${queryArr.length - 1 === i ? 'querySelector-item-selected' : ''}">
                ${idx}:${v}</a>`
            }
            querySelector.innerHTML = b;
            querySelector.classList.remove('hidden');
        } else {
            querySelector.innerHTML = "";
            querySelector.classList.add('hidden');
        }

        navSpace.style.height = `${nav.offsetHeight + 20}px`;
        const p = urlParams.toString();
        for (let i = 0; i < dicts.length; i++) {
            const n = dicts[i].getAttribute('data-dict-name');
            dicts[i].href = `/${n}?${p}`;
        }

        document.title = `${selectedDictAr}: ${word}`;
        window.history.replaceState(null, '', `${window.location.pathname}?${p}`);
    }, 250);
}