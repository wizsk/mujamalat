// nav.js
const selectedDictIdName = "sw-dict-item-selected";
for (let i = 0; i < dicts.length; i++) {
    dicts[i].onclick = async (e) => {
        e.preventDefault();

        const cur = e.target.getAttribute('data-dict-name');
        if (selectedDict === cur) return;
        contentHolder.innerHTML = "<center>انتطر...</center>"
        document.getElementById(selectedDictIdName).id = "";
        e.target.id = selectedDictIdName;
        selectedDict = cur;
        selectedDictAr = e.target.getAttribute('data-dict-name-ar');

        console.log("selected dict:", cur)
        if (cur === "ar_en") {
            ar_en_style.disabled = false;
            eng_style.disabled = true;
        } else if (cur === "hanswehr" || cur === "lanelexcon") {
            eng_style.disabled = false;
            ar_en_style.disabled = true;
        } else {
            eng_style.disabled = true;
            ar_en_style.disabled = true;
        }

        let p = "";
        if (preQuery !== "") {
            p = `?w=${preQuery}&idx=${queryIdx}`;
        }
        document.title = `${selectedDictAr}${currWord !== "" ? ": " + currWord : ""}`;
        window.history.replaceState(null, '', `/${selectedDict}${p}`);

        if (currWord === "") { return };
        console.log(`req: /content?dict=${selectedDict}&w=${currWord}`);
        const r = await fetch(`/content?dict=${selectedDict}&w=${currWord}`).catch((err) =>
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
    }
}