$(document).ready(function () {
  // Initialize Lucide icons
  lucide.createIcons();

  // Splash screen functionality
  setTimeout(function () {
    $("#splash-screen").fadeOut(500, function () {
      $("#main-content").fadeIn(500);
    });
  }, 2000);

  // Service request functionality - redirect to login
  function handleServiceRequest() {
    window.location.href = "/login";
  }

  // Bind click events to CTA buttons
  $("#hero-cta, #cta-button").click(function () {
    handleServiceRequest();
  });

  // Smooth scrolling for navigation links
  $('a[href^="#"]').click(function (e) {
    e.preventDefault();
    var target = $($(this).attr("href"));
    if (target.length) {
      $("html, body").animate(
        {
          scrollTop: target.offset().top - 70,
        },
        800
      );
    }
  });

  // Add hover effects to service cards
  $(".service-card").hover(
    function () {
      $(this).addClass("shadow-lg");
    },
    function () {
      $(this).removeClass("shadow-lg");
    }
  );

  // Add scroll effect to topbar
  $(window).scroll(function () {
    if ($(this).scrollTop() > 50) {
      $(".topbar").addClass("shadow-sm");
    } else {
      $(".topbar").removeClass("shadow-sm");
    }
  });
});
