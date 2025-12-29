const noteWord = document.getElementById("noteWord");
const infoTxt = document.getElementById("infoTxt");
const infoEditBtn = document.getElementById("infoEditBtn");
const infoCloseBtn = document.getElementById("infoCloseBtn");
const noteDelBtn = document.getElementById("noteDelBtn");

const infoEditBtnTxtEdit = "Edit";
const infoEditBtnTxtSave = "Save";
const infoEditBtnTxtWait = "Wait...";

let infoIsEditing = false;
let oldInfo = "";
let noteIsInReq = false;

/** @param {string} word */
async function showInfoModal(word, callBack) {
  if (noteIsInReq) return;

  noteWord.textContent = word;
  infoIsEditing = false;
  infoEditBtn.textContent = infoEditBtnTxtWait;
  infoCloseBtn.disabled = true;

  infoTxt.readOnly = true;
  infoTxt.setAttribute("tabindex", "-1");

  infoDialog.showModal();

  noteIsInReq = true;
  const res = await fetch(`/rd/high_info/${word}`).catch((err) =>
    console.error(err),
  );
  if (res.ok && res.status == 202) {
    infoTxt.value = await res.text();
  } else {
    infoTxt.value = "";
  }
  noteIsInReq = false;
  infoTxt.blur();

  infoEditBtn.textContent = infoEditBtnTxtEdit;
  infoCloseBtn.disabled = false;

  // when deleting just keep the note empty
  infoEditBtn.onclick = async () => {
    if (infoIsEditing) {
      if (noteIsInReq) return;

      infoEditBtn.textContent = infoEditBtnTxtWait;
      noteIsInReq = true;
      infoCloseBtn.disabled = true;
      const val = infoTxt.value.trim();
      const res = await fetch(
        `/rd/high_info/${word}?note=${encodeURIComponent(val)}`,
        { method: "POST" },
      ).catch((err) => console.error(err));

      if (res.status == 202) {
        if (callBack) callBack(val == "");
      } else {
        alert("Could not save note");
      }

      noteIsInReq = false;
      infoCloseBtn.disabled = false;

      infoIsEditing = false;
      infoTxt.readOnly = true;
      infoTxt.setAttribute("tabindex", "-1");
      infoEditBtn.textContent = infoEditBtnTxtEdit;
      return;
    }

    infoEditBtn.textContent = infoEditBtnTxtSave;
    infoIsEditing = true;
    oldInfo = infoTxt.value.trim() ? infoTxt.value.trim() : "";
    infoTxt.readOnly = false;
    infoTxt.removeAttribute("tabindex");
    infoTxt.focus();
  };

  noteDelBtn.onclick = async () => {
    if (!confirm("Do you really want to delete note?") && noteIsInReq) return;
    noteIsInReq = true;
    const res = await fetch(`/rd/high_info/${word}`, { method: "POST" }).catch(
      (err) => console.error(err),
    );

    if (res.ok || res.status == 202) {
      if (callBack) callBack(true);
      infoTxt.value = "";
    } else {
      alert("Could not delete note");
    }

    noteIsInReq = false;
  };
}

infoDialog.addEventListener("close", (e) => {
  if (infoIsEditing && !confirm("Do you really want to close?")) {
    infoDialog.showModal();
  }
});

infoCloseBtn.onclick = () => {
  infoDialog.close();
};
