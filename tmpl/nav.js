// nav.js
const selectedDictIdName = "sw-dict-item-selected";
for (let i = 0; i < dicts.length; i++) {
    dicts[i].onclick = async (e) => {
        e.preventDefault();
        console.log(e)

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

        // {{if not .RDMode}}
        let p = "";
        if (preQuery !== "") {
            p = `?w=${preQuery}&idx=${queryIdx}`;
        }
        document.title = `${selectedDictAr}${currWord !== "" ? ": " + currWord : ""}`;
        window.history.replaceState(null, '', `/${selectedDict}${p}`);
        // {{end}}

        getResAndShow(currWord);
    }
}