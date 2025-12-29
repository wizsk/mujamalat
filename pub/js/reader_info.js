const noteWord = document.getElementById("noteWord");
const noteTxtAr = document.getElementById("infoTxt");
const noteEditBtn = document.getElementById("infoEditBtn");
const noteCloseBtn = document.getElementById("infoCloseBtn");
const noteDelBtn = document.getElementById("noteDelBtn");

const noteEditBtnTxtEdit = "Edit";
const noteEditBtnTxtSave = "Save";
const noteEditBtnTxtWait = "Wait…";

let infoIsEditing = false;
let oldNote = "";

/** @param {string} word */
async function showInfoModal(word, callBack) {
  // book keeping
  noteWord.textContent = word;
  noteTxtAr.value = "";

  infoIsEditing = false;

  noteEditBtn.textContent = noteEditBtnTxtWait;
  noteEditBtn.disabled = true;

  noteTxtAr.readOnly = true;
  noteTxtAr.setAttribute("tabindex", "-1");

  infoDialog.showModal();
  noteTxtAr.blur();

  const res = await fetch(`/rd/high_info/${word}`).catch((err) => {
    console.error(err);
    alert("Something went wrong. Could not fetch not");
  });

  if (res.ok && res.status == 202) {
    const val = await res.text();
    // if there is no note start in editing mode
    if (val) {
      noteEditBtn.textContent = noteEditBtnTxtEdit;
    } else {
      infoIsEditing = true;
      noteEditBtn.textContent = noteEditBtnTxtSave;
      noteTxtAr.readOnly = false;
      noteTxtAr.removeAttribute("tabindex");
      noteTxtAr.focus();
    }
    noteEditBtn.disabled = false;
    noteTxtAr.value = val;
    oldNote = val;
  } else {
    alert("Something went wrong. Could not fetch note");
  }


  // when deleting just keep the note empty
  noteEditBtn.onclick = async () => {
    // was in editing mode now save and exit it
    if (infoIsEditing) {
      noteEditBtn.textContent = noteEditBtnTxtWait;
      noteEditBtn.disabled = true;

      const val = noteTxtAr.value.trim();
      const res = await fetch(
        `/rd/high_info/${word}?note=${encodeURIComponent(val)}`,
        { method: "POST" },
      ).catch((err) => console.error(err));

      if (res.status == 202) {
        if (callBack) callBack(val == "");
      } else {
        alert("Could not save note");
      }

      noteEditBtn.disabled = false;
      noteTxtAr.readOnly = true;
      noteTxtAr.setAttribute("tabindex", "-1");
      noteEditBtn.textContent = noteEditBtnTxtEdit;

      infoIsEditing = false;
      return;
    }

    // enable editing mode
    noteEditBtn.textContent = noteEditBtnTxtSave;
    noteTxtAr.readOnly = false;
    noteTxtAr.removeAttribute("tabindex");
    noteTxtAr.focus();

    infoIsEditing = true;
  };

  noteDelBtn.onclick = async () => {
    if (!confirm("Do you really want to delete note?")) return;

    noteDelBtn.disabled = true;
    noteEditBtn.disabled = true;
    noteEditBtn.textContent = noteEditBtnTxtEdit;
    infoIsEditing = false;

    const res = await fetch(`/rd/high_info/${word}`, { method: "POST" }).catch(
      (err) => console.error(err),
    );

    if (res.ok || res.status == 202) {
      noteTxtAr.value = "";
      if (callBack) callBack(true);
    } else {
      alert("Could not delete note");
    }

    noteDelBtn.disabled = false;
    noteEditBtn.disabled = false;
  };
}

infoDialog.addEventListener("close", (e) => {
  if (oldNote != noteTxtAr.value && !confirm("Do you really want to close?")) {
    infoDialog.showModal();
  }
});

// infoCloseBtn.onclick = () => {
//   infoDialog.close();
// };
