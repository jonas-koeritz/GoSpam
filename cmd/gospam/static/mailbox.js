$('.accordion.content').hide();

$('.accordion.title').on("click", function(e) {
  titleRow = $(e.target.closest('tr'))
  contentRow = $(e.target.closest('tr')).next('tr')
  if (contentRow.is(":visible")) {
    titleRow.find('.icon').removeClass("expanded");
    contentRow.find('.content').slideUp("slow", function() {
      contentRow.hide();
    });
  } else {
    titleRow.find('.icon').addClass("expanded");
    contentRow.show(0, function() {
      contentRow.find('.content').slideDown("slow");
    });
  }
});