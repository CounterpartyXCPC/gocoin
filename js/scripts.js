/* Template: Riga - Mobile App Landing Page Template
   Author: InovatikThemes
   Created: Apr 2018
   Description: Custom JS file
*/

// Snake Waves

class App {
  init() {
    this.backgroundColor = '#6a2bff';
    this.spotLightColor = 0xffffff;
    this.angle = 0;
    this.spheres = [];
    this.holes = [];
    this.gui = new dat.GUI();

    this.velocity = .08;
    this.amplitude = 5;
    this.waveLength = 20;

    this.scene = new THREE.Scene();

    this.camera = new THREE.PerspectiveCamera(20, window.innerWidth / window.innerHeight, 1, 1000);
    this.camera.position.set(60, 60, -60);

    this.addRenderer();

    this.controls = new THREE.OrbitControls(this.camera, this.renderer.domElement);

    this.addAmbientLight();

    this.addSpotLight();

    const backgroundGUI = this.gui.addFolder('Background');
    backgroundGUI.addColor(this, 'backgroundColor').onChange((color) => {
      document.body.style.backgroundColor = color;
    });

    const obj = {
      color: '#ffffff',
      emissive: '#e07cff',
      reflectivity: 1,
      metalness: .2,
      roughness: 0
    };

    const material = new THREE.MeshPhysicalMaterial(obj);
    const geometry = new THREE.SphereGeometry(.5, 32, 32);

    const tileTop = { color: '#fa3fce' };
    const tileTopMaterial = new THREE.MeshBasicMaterial(tileTop);

    const tileInside = { color: '#671c87' };
    const tileInsideMaterial = new THREE.MeshBasicMaterial(tileInside);

    const materials = [tileTopMaterial, tileInsideMaterial];
    const props = {
      steps: 1,
      depth: 1,
      bevelEnabled: false
    };

    const guiWave = this.gui.addFolder('Wave');
    guiWave.add(this, 'waveLength', 0, 20).onChange((waveLength) => {
      this.waveLength = waveLength;
    });

    guiWave.add(this, 'amplitude', 3, 10).onChange((amplitude) => {
      this.amplitude = amplitude;
    });

    guiWave.add(this, 'velocity', 0, .2).onChange((velocity) => {
      this.velocity = velocity;
    });


    this.createSet(1, 1, geometry, material, props, materials);

    this.createSet(4, 1, geometry, material, props, materials);

    this.createSet(7, 1, geometry, material, props, materials);

    this.createSet(10, 1, geometry, material, props, materials);

    this.createSet(-2, 1, geometry, material, props, materials);

    this.createSet(-5, 1, geometry, material, props, materials);

    this.createSet(-8, 1, geometry, material, props, materials);

    this.createSet(-11, 1, geometry, material, props, materials);

    this.addFloorShadow();

    this.animate();

    window.addEventListener('resize', this.onResize.bind(this));
  }

  radians(degrees) {
    return degrees * Math.PI / 180;
  }

  createSet(x, z, geometry, material, props, materials) {
    this.floorShape = this.createShape();

    this.createRingOfHoles(this.floorShape, 1, 0);

    const geometryTile = new THREE.ExtrudeGeometry(this.floorShape, props);

    this.createGround(this.floorShape, x, z, geometryTile, materials);

    this.addSphere(x, z, geometry, material);
  }

  createRingOfHoles(shape, count, radius) {
    const l = 360 / count;
    const distance = (radius * 2);

    for (let index = 0; index < count; index++) {
      const pos = this.radians(l * index);
      const sin = Math.sin(pos) * distance;
      const cos = Math.cos(pos) * distance;

      this.createHoles(shape, sin, cos);
    }
  }

  createShape() {
    const size = 1;
    const vectors = [
      new THREE.Vector2(-size, size),
      new THREE.Vector2(-size, -size),
      new THREE.Vector2(size, -size),
      new THREE.Vector2(size, size)
    ];

    const shape = new THREE.Shape(vectors);

    shape.autoClose = true;

    return shape;
  }

  createHoles(shape, x, z) {
    const radius = .5;
    const holePath = new THREE.Path();

    holePath.moveTo(x, z);
    holePath.ellipse(x, z, radius, radius, 0, Math.PI * 2);

    holePath.autoClose = true;

    shape.holes.push(holePath);

    this.holes.push({
      x,
      z
    });
  }

  addFloorShadow() {
    const planeGeometry = new THREE.PlaneGeometry(50, 50);
    const planeMaterial = new THREE.ShadowMaterial({ opacity: .08 });

    this.floor = new THREE.Mesh(planeGeometry, planeMaterial);

    planeGeometry.rotateX(- Math.PI / 2);

    this.floor.position.y = -10;
    this.floor.receiveShadow = true;

    this.scene.add(this.floor);
  }

  createGround(shape, x, z, geometry, materials) {
    const mesh = new THREE.Mesh(geometry, materials);

    mesh.needsUpdate = true;

    mesh.rotation.set(Math.PI * 0.5, 0, 0);

    mesh.position.set(x, 0, z);

    this.scene.add(mesh);
  }

  hexToRgbTreeJs(hex) {
    const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);

    return result ? {
      r: parseInt(result[1], 16) / 255,
      g: parseInt(result[2], 16) / 255,
      b: parseInt(result[3], 16) / 255
    } : null;
  }

  addAmbientLight() {
    this.ambientLight = new THREE.AmbientLight(0x6e6e6e, 1);
    this.scene.add(this.ambientLight);
  }

  addSpotLight() {
    this.spotLight = new THREE.SpotLight(0xffffff);
    this.spotLight.position.set(0, 30, 0);
    this.spotLight.castShadow = true;

    this.scene.add(this.spotLight);
  }

  addRenderer() {
    this.renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true });
    this.renderer.shadowMap.enabled = true;
    this.renderer.shadowMap.type = THREE.PCFSoftShadowMap;
    this.renderer.setSize(window.innerWidth, window.innerHeight);

    document.body.appendChild(this.renderer.domElement);
  }

  addSphere(x, z, geometry, material) {
    const mesh = new THREE.Mesh(geometry, material);
    mesh.position.set(x, 2, z);
    mesh.castShadow = true;
    mesh.receiveShadow = true;

    this.spheres.push(mesh);

    this.scene.add(mesh);
  }

  distance(x1, y1, x2, y2) {
    return Math.sqrt(Math.pow((x1 - x2), 2) + Math.pow((y1 - y2), 2));
  }

  map(value, start1, stop1, start2, stop2) {
    return (value - start1) / (stop1 - start1) * (stop2 - start2) + start2
  }

  drawWave() {
    const total = this.spheres.length;

    for (let i = 0; i < total; i++) {
      const distance = this.distance(this.spheres[i].position.z, this.spheres[i].position.x, 100, 100);

      const offset = this.map(distance, 0, 100, this.waveLength, -this.waveLength);

      const angle = this.angle + offset;

      const y = this.map(Math.sin(angle), -1, 1, -3, this.amplitude);

      this.spheres[i].position.y = y;
    }

    this.angle += this.velocity;
  }

  animate() {
    this.drawWave();

    this.controls.update();

    this.renderer.render(this.scene, this.camera);

    requestAnimationFrame(this.animate.bind(this));
  }

  onResize() {
    const ww = window.innerWidth;
    const wh = window.innerHeight;

    this.camera.aspect = ww / wh;
    this.camera.updateProjectionMatrix();
    this.renderer.setSize(ww, wh);
  }
}

new App().init();

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
	$('#clock').countdown('2018/11/15 16:40:00', "GMT") /* change here your "countdown to" date */
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