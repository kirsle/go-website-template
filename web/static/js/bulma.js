// Hamburger menu script for mobile.
document.addEventListener('DOMContentLoaded', () => {

  // Make all off-site hyperlinks open in a new tab.
  (document.querySelectorAll("a") || []).forEach(node => {
    let href = node.attributes.href;
    if (href === undefined) return;
    href = href.textContent;
    if (href.indexOf("http:") === 0 || href.indexOf("https:") === 0) {
      node.target = "_blank";
    }
  });

  // Hamburger menu script.
  (function() {
    // Get all "navbar-burger" elements
    const $navbarBurgers = Array.prototype.slice.call(document.querySelectorAll('.navbar-burger'), 0);

    // Add a click event on each of them
    $navbarBurgers.forEach( el => {
      el.addEventListener('click', () => {

        // Get the target from the "data-target" attribute
        const target = el.dataset.target;
        const $target = document.getElementById(target);

        // Toggle the "is-active" class on both the "navbar-burger" and the "navbar-menu"
        el.classList.toggle('is-active');
        $target.classList.toggle('is-active');

      });
    });
  })();

  // Allow the "More" drop-down to work on mobile (toggle is-active on click instead of requiring mouseover)
  (function() {
    const menu = document.querySelector("#navbar-more"),
      userMenu = document.querySelector("#navbar-user"),
      activeClass = "is-active";

    if (!menu) return;

    // Click the "More" menu to permanently toggle the menu.
    menu.addEventListener("click", (e) => {
      if (menu.classList.contains(activeClass)) {
        menu.classList.remove(activeClass);
      } else {
        menu.classList.add(activeClass);
      }
      e.stopPropagation();
    });

    // Touching the user drop-down button toggles it.
    if (userMenu !== null) {
      userMenu.addEventListener("touchstart", (e) => {
        // On mobile/tablet screens they had to hamburger menu their way here anyway, let it thru.
        if (screen.width < 1024) {
          return;
        }

        e.preventDefault();
        if (userMenu.classList.contains(activeClass)) {
          userMenu.classList.remove(activeClass);
        } else {
          userMenu.classList.add(activeClass);
        }
      });
    }

    // Touching a link from the user menu lets it click thru.
    (document.querySelectorAll(".navbar-dropdown") || []).forEach(node => {
        node.addEventListener("touchstart", (e) => {
          e.stopPropagation();
        });
    });

    // Clicking the rest of the body will close an active navbar-dropdown.
    (document.addEventListener("click", (e) => {
      (document.querySelectorAll(".navbar-dropdown.is-active, .navbar-item.is-active") || []).forEach(node => {
        node.classList.remove(activeClass);
      });
    }));
  })();

  // Common event handlers for bulma modals.
  (document.querySelectorAll(".modal-background, .modal-close, .photo-modal") || []).forEach(node => {
    const target = node.closest(".modal");
    node.addEventListener("click", () => {
      target.classList.remove("is-active");
    });
  });
});