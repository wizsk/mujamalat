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

/**
 *
 * @param {string} word
 * @param {Function} callBack - when calleing it send a bool if del then true
*/
async function showInfoModal(word, callBack) {
  // book keeping
  noteWord.textContent = word;
  noteTxtAr.value = "";

  infoIsEditing = false;

  noteEditBtn.textContent = noteEditBtnTxtWait;
  noteEditBtn.disabled = true;

  noteTxtAr.readOnly = true;

  infoDialog.showModal();
  noteTxtAr.focus();

  const res = await fetch(`/rd/high_info/${word}`).catch((err) => {
    console.error(err);
    alert("Something went wrong. Could not fetch note");
  });

  if (res.ok && res.status == 202) {
    const val = await res.text();
    noteTxtAr.value = val;
    oldNote = val;
    // if there is no note start in editing mode
    if (val) {
      noteEditBtn.textContent = noteEditBtnTxtEdit;
    } else {
      infoIsEditing = true;
      noteEditBtn.textContent = noteEditBtnTxtSave;
      noteTxtAr.readOnly = false;
      noteTxtAr.focus();
    }
    noteEditBtn.disabled = false;
  } else {
    alert("Something went wrong. Could not fetch note");
  }


  // when deleting just keep the note empty
  noteEditBtn.onclick = async () => {
    // was in editing mode now save and exit it
    if (infoIsEditing) {
      noteEditBtn.disabled = true;
      noteEditBtn.textContent = noteEditBtnTxtWait;

      noteTxtAr.readOnly = true;
      const val = noteTxtAr.value.trim();
      noteTxtAr.value = val;

      const res = await fetch(
        `/rd/high_info/${word}?note=${encodeURIComponent(val)}`,
        { method: "POST" },
      ).catch((err) => console.error(err));

      if (res.status == 202) {
        oldNote = val;
        if (callBack) callBack(val == "");
      } else {
        alert("Could not save note");
      }

      noteEditBtn.disabled = false;
      noteEditBtn.textContent = noteEditBtnTxtEdit;

      infoIsEditing = false;
      return;
    }

    // enable editing mode
    noteEditBtn.textContent = noteEditBtnTxtSave;
    noteTxtAr.readOnly = false;
    noteTxtAr.focus();

    infoIsEditing = true;
  };

  noteDelBtn.onclick = async () => {
    if (!confirm("Do you really want to delete note?")) return;

    noteEditBtn.disabled = true;
    noteDelBtn.disabled = true;
    noteTxtAr.readOnly = true;

    noteEditBtn.textContent = noteEditBtnTxtEdit;
    infoIsEditing = false;

    const res = await fetch(`/rd/high_info/${word}`, { method: "POST" }).catch(
      (err) => console.error(err),
    );

    if (res.ok || res.status == 202) {
      noteTxtAr.value = "";
      oldNote = "";
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

function infoModalClose() {
  infoDialog.close();
}

noteTxtAr.addEventListener("keydown", (e) => {
  if (!e.ctrlKey || e.shiftKey) return;

  switch (e.code) {
      case "KeyI":
        e.preventDefault();
        if (!infoIsEditing) noteEditBtn.click();
        noteTxtAr.focus();
        break;

    case "Enter":
      e.preventDefault();
      if (infoIsEditing) noteEditBtn.click();
      break;

  }
})

// infoCloseBtn.onclick = () => {
//   infoDialog.close();
// };
