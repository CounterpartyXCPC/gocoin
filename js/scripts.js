/* Template: Riga - Mobile App Landing Page Template
   Author: InovatikThemes
   Created: Apr 2018
   Description: Custom JS file
*/


(function($) {
    "use strict"; 
	
	/* Preloader */
	$(window).on('load', function() {
		var preloaderFadeOutTime = 500;
		function hidePreloader() {
			var preloader = $('.spinner-wrapper');
			setTimeout(function() {
				preloader.fadeOut(preloaderFadeOutTime);
			}, 500);
		}
		hidePreloader();
	});

	
	/* Navbar Scripts */
	// jQuery for page scrolling feature - requires jQuery Easing plugin
	$(function() {
		$(document).on('click', 'a.page-scroll', function(event) {
			var $anchor = $(this);
			$('html, body').stop().animate({
				scrollTop: $($anchor.attr('href')).offset().top
			}, 600, 'easeInOutExpo');
			$('input:checkbox').prop('checked', false); // closes the responsive menu on menu item click
			event.preventDefault();
		});
	});


	/* Countdown Timer */
	$('#clock').countdown('2018/12/27 08:50:56') /* change here your "countdown to" date */
	.on('update.countdown', function(event) {
		var format = '<span class="counter-number">%D<br><span class="timer-text">Days</span></span><span class="separator">:</span><span class="counter-number">%H<br><span class="timer-text">Hours</span></span><span class="separator">:</span><span class="counter-number">%M<br><span class="timer-text">Minutes</span></span><span class="separator">:</span><span class="counter-number">%S<br><span class="timer-text">Seconds</span></span>';
			
		$(this).html(event.strftime(format));
	})
	.on('finish.countdown', function(event) {
	$(this).html('This offer has expired!')
		.parent().addClass('disabled');
	});


	/* Signup Form */
    $("#SignupForm").validator().on("submit", function(event) {
    	if (event.isDefaultPrevented()) {
            // handle the invalid form...
            formError();
            submitMSG(false, "Check if all fields are filled in!");
        } else {
            // everything looks good!
            event.preventDefault();
            submitForm();
        }
    });

    function submitForm() {
        // initiate variables with form content
		var email = $("#email").val();
        
        $.ajax({
            type: "POST",
            url: "php/signupform-process.php",
            data: "email=" + email, 
            success: function(text) {
                if (text == "success") {
                    formSuccess();
                } else {
                    formError();
                    submitMSG(false, text);
                }
            }
        });
	}

    function formSuccess() {
        $("#SignupForm")[0].reset();
        submitMSG(true, "Email Submitted!")
    }

    function formError() {
        $("#SignupForm").removeClass().addClass('shake animated').one('webkitAnimationEnd mozAnimationEnd MSAnimationEnd oanimationend animationend', function() {
            $(this).removeClass();
        });
	}

    function submitMSG(valid, msg) {
        if (valid) {
            var msgClasses = "h3 text-confirmation";
        } else {
            var msgClasses = "h3 text-error";
        }
        $("#msgSubmit").removeClass().addClass(msgClasses).text(msg);
	}


	/* Cards Slider Swiper */
	var cardsSlider = new Swiper('.cardsSlider', {
		autoplay: {
			delay: 4000,
		},
        loop: false,
        pagination: {
			el: '.swiper-pagination',
			type: 'bullets',
			clickable: true,
		},
		slidesPerView: 4,
		spaceBetween: 0,
        breakpoints: {
			1200: {
				slidesPerView: 3,
				spaceBetween: 30,
			},
			992: {
				slidesPerView: 2,
				spaceBetween: 30,
			},
			540: {
				slidesPerView: 1,
				spaceBetween: 20,
			}
		}
	});


	/* Image Slider Gallery Swiper */
	var imageSliderGallery = new Swiper('.imageSliderGallery', {
		autoplay: {
			delay: 3000,
		},
        loop: false,
		slidesPerView: 6,
		spaceBetween: 30,
		navigation: {
			nextEl: '.swiper-button-next',
			prevEl: '.swiper-button-prev',
		},
		breakpoints: {
			992: {
				slidesPerView: 3,
				spaceBetween: 30
			},
			768: {
				slidesPerView: 2,
				spaceBetween: 20
			},
			576: {
				slidesPerView: 1,
				spaceBetween: 10
			}
		}
	});


	/* Image Slider Gallery Magnific Popup */
	$('.popup-link').magnificPopup({
		fixedContentPos: false, /* keep it false to avoid html tag shift with margin-right: 17px */
		removalDelay: 200,
		type: 'image',
		callbacks: {
			beforeOpen: function() {
				this.st.image.markup = this.st.image.markup.replace('mfp-figure', 'mfp-figure ' + this.st.el.attr('data-effect'));
			}
		},
		gallery:{
			enabled:true //enable gallery mode
		}
	});


	/* Lightbox Magnific PopUp */
	$('.popup-with-move-anim').magnificPopup({
		type: 'inline',
		fixedContentPos: false, /* keep it false to avoid html tag shift with margin-right: 17px */
		fixedBgPos: true,
		overflowY: 'auto',
		closeBtnInside: true,
		preloader: false,
		midClick: true,
		removalDelay: 200,
		mainClass: 'my-mfp-slide-bottom'
	});


	/* Statistics Numbers CountTo */
	var a = 0;
	$(window).scroll(function() {
		if ($('#counter').length) { // checking if CountTo section exists in the page, if not it will not run the script and avoid errors	
			var oTop = $('#counter').offset().top - window.innerHeight;
			if (a == 0 && $(window).scrollTop() > oTop) {
			$('.counter-value').each(function() {
				var $this = $(this),
				countTo = $this.attr('data-count');
				$({
				countNum: $this.text()
				}).animate({
					countNum: countTo
				},
				{
					duration: 2000,
					easing: 'swing',
					step: function() {
					$this.text(Math.floor(this.countNum));
					},
					complete: function() {
					$this.text(this.countNum);
					//alert('finished');
					}
				});
			});
			a = 1;
			}
		}
	});
	
	
	/* Contact Form */
    $("#ContactForm").validator().on("submit", function(event) {
    	if (event.isDefaultPrevented()) {
            // handle the invalid form...
            formCError();
            submitCMSG(false, "Please fill all fields!");
        } else {
            // everything looks good!
            event.preventDefault();
            submitCForm();
        }
    });

    function submitCForm() {
        // initiate variables with form content
		var name = $("#cname").val();
		var email = $("#cemail").val();
        var message = $("#cmessage").val();
        
        $.ajax({
            type: "POST",
            url: "php/contactform-process.php",
            data: "name=" + name + "&email=" + email + "&message=" + message, 
            success: function(text) {
                if (text == "success") {
                    formCSuccess();
                } else {
                    formCError();
                    submitCMSG(false, text);
                }
            }
        });
	}

    function formCSuccess() {
        $("#ContactForm")[0].reset();
        submitCMSG(true, "Message Submitted!")
    }

    function formCError() {
        $("#ContactForm").removeClass().addClass('shake animated').one('webkitAnimationEnd mozAnimationEnd MSAnimationEnd oanimationend animationend', function() {
            $(this).removeClass();
        });
	}

    function submitCMSG(valid, msg) {
        if (valid) {
            var msgClasses = "h3 text-center tada animated";
        } else {
            var msgClasses = "h3 text-center";
        }
        $("#msgCSubmit").removeClass().addClass(msgClasses).text(msg);
	}


	/* Back To Top Button */
    // create the back to top button
    $('body').prepend('<a href="body" class="back-to-top page-scroll">Back to Top</a>');
    var amountScrolled = 700;
    $(window).scroll(function() {
        if ($(window).scrollTop() > amountScrolled) {
            $('a.back-to-top').fadeIn('500');
        } else {
            $('a.back-to-top').fadeOut('500');
        }
    });


	/* Removes Long Focus On Buttons */
	$(".button, a, button").mouseup(function() {
		$(this).blur();
	});


})(jQuery);