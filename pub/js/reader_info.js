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

/** @param {string} word */
async function showInfoModal(word, callBack) {
  // book keeping
  noteWord.textContent = word;
  infoTxt.value = "";

  infoIsEditing = false;

  infoEditBtn.textContent = infoEditBtnTxtWait;
  infoEditBtn.disabled = true;

  infoTxt.readOnly = true;
  infoTxt.setAttribute("tabindex", "-1");

  infoDialog.showModal();
  infoTxt.blur();

  const res = await fetch(`/rd/high_info/${word}`).catch((err) => {
    console.error(err);
    alert("Something went wrong. Could not fetch not");
  });

  if (res.ok && res.status == 202) {
    const val = await res.text();
    // if there is no note start in editing mode
    if (val) {
      infoEditBtn.textContent = infoEditBtnTxtEdit;
    } else {
      infoIsEditing = true;
      infoEditBtn.textContent = infoEditBtnTxtSave;
      infoTxt.readOnly = false;
      infoTxt.removeAttribute("tabindex");
      infoTxt.focus();
    }
    infoEditBtn.disabled = false;
    infoTxt.value = val;
  } else {
    alert("Something went wrong. Could not fetch note");
  }


  // when deleting just keep the note empty
  infoEditBtn.onclick = async () => {
    // was in editing mode now save and exit it
    if (infoIsEditing) {
      infoEditBtn.textContent = infoEditBtnTxtWait;
      infoEditBtn.disabled = true;

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

      infoEditBtn.disabled = false;
      infoTxt.readOnly = true;
      infoTxt.setAttribute("tabindex", "-1");
      infoEditBtn.textContent = infoEditBtnTxtEdit;

      infoIsEditing = false;
      return;
    }

    // enable editing mode
    infoEditBtn.textContent = infoEditBtnTxtSave;
    oldInfo = infoTxt.value.trim() ? infoTxt.value.trim() : "";
    infoTxt.readOnly = false;
    infoTxt.removeAttribute("tabindex");
    infoTxt.focus();

    infoIsEditing = true;
  };

  noteDelBtn.onclick = async () => {
    if (!confirm("Do you really want to delete note?")) return;

    noteDelBtn.disabled = true;
    infoEditBtn.disabled = true;
    infoEditBtn.textContent = infoEditBtnTxtEdit;
    infoIsEditing = false;

    const res = await fetch(`/rd/high_info/${word}`, { method: "POST" }).catch(
      (err) => console.error(err),
    );

    if (res.ok || res.status == 202) {
      infoTxt.value = "";
      if (callBack) callBack(true);
    } else {
      alert("Could not delete note");
    }

    noteDelBtn.disabled = false;
    infoEditBtn.disabled = false;
  };
}

infoDialog.addEventListener("close", (e) => {
  if (infoIsEditing && !confirm("Do you really want to close?")) {
    infoDialog.showModal();
  }
});

// infoCloseBtn.onclick = () => {
//   infoDialog.close();
// };
