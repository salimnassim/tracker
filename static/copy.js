(function () {
  document.addEventListener('click', function (event) {
    var button = event.target.closest('[data-copy]');
    if (!button) {
      return;
    }

    var value = button.getAttribute('data-copy-value');
    if (!value || !navigator.clipboard || !navigator.clipboard.writeText) {
      return;
    }

    navigator.clipboard.writeText(value).then(function () {
      var originalLabel = button.textContent;
      clearTimeout(button._timer);
      button.textContent = 'Copied';
      button._timer = setTimeout(function () {
        button.textContent = originalLabel;
      }, 1500);
    }).catch(function () {});
  });
})();
