<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Separator } from '$lib/components/ui/separator';
	import { Carousel, CarouselContent, CarouselItem, CarouselNext, CarouselPrevious } from '$lib/components/ui/carousel';
	import { 
		CheckCircle, 
		Calendar, 
		Smartphone, 
		ArrowRight,
		Play,
		Star,
		Target,
		Dock,
		Terminal
	} from 'lucide-svelte';

	const features = [
		{
			icon: Calendar,
			title: 'Weekly Planning',
			description: 'Organize your tasks by day of the week with beautiful, intuitive cards'
		},
		{
			icon: Smartphone,
			title: 'Mobile Friendly',
			description: 'PWA support with touch gestures and responsive design for any device'
		},
		{
			icon: CheckCircle,
			title: 'Simple & Clean',
			description: 'Minimalist interface that focuses on what matters most'
		},
		{
			icon: Target,
			title: 'Tag-Based Filtering',
			description: 'Organize and filter tasks with powerful tag system'
		}
	];



	const screenshots = [
		{
			src: '/week-card-layout.png',
			alt: 'Zendo Weekly View',
			title: 'Weekly Planning',
			description: 'Organize tasks by day with beautiful color-coded cards'
		},
		{
			src: '/todays-tasks.png',
			alt: 'Zendo Today\'s Tasks',
			title: 'Today\'s Focus',
			description: 'Focus on what matters today with a clean, distraction-free view'
		},
		{
			src: '/weeks-tasks.png',
			alt: 'Zendo Week\'s Tasks',
			title: 'Week Overview',
			description: 'See all your week\'s tasks in one organized, filterable view'
		},
        {
			src: '/jump-weeks-by-calendar.png',
			alt: 'Zendo Calendar View',
			title: 'Calendar Selection',
			description: 'Navigate between weeks with an NLP dates or calendar interface'
		},
		{
			src: '/filter-by-tags.png',
			alt: 'Zendo Tag Filtering',
			title: 'Tag Filtering',
			description: 'Filter and organize tasks with tag-based system'
		}
	];

	const dockerComposeYaml = `
services:
  zendo:
    image: ghcr.io/rishikanthc/zendo:v0.2.0
    container_name: zendo
    platform: linux/amd64
    ports:
      - "8080:8080"
    environment:
      - TZ=\${TZ:-America/Los_Angeles}
    volumes:
      - ./storage:/app/storage
    restart: unless-stopped`;

	function scrollToSection(sectionId: string) {
		const element = document.getElementById(sectionId);
		if (element) {
			element.scrollIntoView({ behavior: 'smooth' });
		}
	}
</script>

<svelte:head>
	<title>Zendo - Simple Weekly Task Management</title>
	<meta name="description" content="A beautiful, minimalist weekly task manager designed for simplicity and focus. Organize your week with elegant day-based planning." />
</svelte:head>

<div class="min-h-screen bg-gradient-to-br from-gray-50 to-gray-100">
	<!-- Navigation -->
	<nav class="border-b bg-white/80 backdrop-blur-sm sticky top-0 z-50">
		<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
			<div class="flex justify-between items-center h-16">
				<div class="flex items-center">
					<h1 class="text-2xl font-megrim uppercase tracking-wider text-gray-900">Zendo</h1>
				</div>
				<div class="flex items-center space-x-4">
					<Button variant="ghost" class="text-gray-600 hover:text-gray-900" onclick={() => scrollToSection('features')}>
						Features
					</Button>
					<Button class="bg-gray-900 hover:bg-gray-800 text-white" onclick={() => scrollToSection('installation')}>
						Get Started
					</Button>
				</div>
			</div>
		</div>
	</nav>

	<!-- Hero Section -->
	<section class="relative overflow-hidden bg-white">
		<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-24">
			<div class="grid lg:grid-cols-2 gap-12 items-center">
				<div class="space-y-8">
					<div class="space-y-4">
						<Badge variant="secondary" class="w-fit">
							<Star class="w-3 h-3 mr-1" />
							Simple Task Management
						</Badge>
						<h1 class="text-5xl lg:text-6xl font-bold text-gray-900 leading-tight">
							Organize Your Week with
							<span class="text-transparent bg-clip-text bg-gradient-to-r from-orange-500 to-red-500 font-megrim">
								Elegant Simplicity
							</span>
						</h1>
						<p class="text-xl text-gray-600 leading-relaxed">
							Zendo transforms weekly planning into a beautiful, mindful experience. 
							With PWA support, tag-based filtering, and a clean interface, organize your week with elegant simplicity.
						</p>
					</div>
					
					<div class="flex flex-col sm:flex-row gap-4">
						<Button size="lg" class="bg-gray-900 hover:bg-gray-800 text-white px-8 py-3" onclick={() => scrollToSection('installation')}>
							<Play class="w-4 h-4 mr-2" />
							Try Zendo
						</Button>
						<Button variant="outline" size="lg" class="px-8 py-3" onclick={() => scrollToSection('features')}>
							<ArrowRight class="w-4 h-4 mr-2" />
							Learn More
						</Button>
					</div>
				</div>

				<div class="relative">
					<div class="relative z-10">
						<Carousel class="w-full max-w-md mx-auto">
							<CarouselContent>
								{#each screenshots as screenshot}
									<CarouselItem>
										<div class="space-y-4">
											<div class="text-center">
												<h3 class="text-lg font-semibold text-gray-900 mb-2">{screenshot.title}</h3>
												<p class="text-sm text-gray-600">{screenshot.description}</p>
											</div>
											<div class="relative">
												<img 
													src={screenshot.src} 
													alt={screenshot.alt} 
													class="w-full rounded-2xl shadow-2xl border border-gray-200"
												/>
											</div>
										</div>
									</CarouselItem>
								{/each}
							</CarouselContent>
							<CarouselPrevious class="left-2" />
							<CarouselNext class="right-2" />
						</Carousel>
					</div>
					<div class="absolute -inset-4 bg-gradient-to-r from-orange-100 to-red-100 rounded-3xl -z-10"></div>
				</div>
			</div>
		</div>
	</section>

	<!-- Features Section -->
	<section id="features" class="py-24 bg-gray-50">
		<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
			<div class="text-center mb-16">
				<h2 class="text-4xl font-bold text-gray-900 mb-4">
					Why Choose Zendo?
				</h2>
				<p class="text-xl text-gray-600 max-w-3xl mx-auto">
					Built for those who believe that the best tools are the ones that get out of your way 
					and help you focus on what truly matters.
				</p>
			</div>

			<div class="grid md:grid-cols-2 lg:grid-cols-4 gap-8">
				{#each features as feature}
					<Card class="border-0 shadow-lg bg-white hover:shadow-xl transition-shadow duration-300">
						<CardHeader class="text-center pb-4">
							<div class="w-12 h-12 bg-gradient-to-br from-orange-100 to-red-100 rounded-xl flex items-center justify-center mx-auto mb-4">
								<svelte:component this={feature.icon} class="w-6 h-6 text-orange-600" />
							</div>
							<CardTitle class="text-lg font-semibold text-gray-900">{feature.title}</CardTitle>
						</CardHeader>
						<CardContent class="text-center">
							<CardDescription class="text-gray-600">{feature.description}</CardDescription>
						</CardContent>
					</Card>
				{/each}
			</div>
		</div>
	</section>

	<!-- Screenshots Section -->
	<section class="py-24 bg-white">
		<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
			<div class="text-center mb-16">
				<h2 class="text-4xl font-bold text-gray-900 mb-4">
					Beautiful in Every View
				</h2>
				<p class="text-xl text-gray-600 max-w-3xl mx-auto">
					Switch between weekly planning, today's focus, this week's overview, and tag filtering with elegant, 
					thoughtfully designed interfaces.
				</p>
			</div>

			<div class="grid lg:grid-cols-2 xl:grid-cols-5 gap-8">
				{#each screenshots as screenshot}
					<div class="space-y-4">
						<div class="text-center">
							<h3 class="text-xl font-semibold text-gray-900 mb-2">{screenshot.title}</h3>
							<p class="text-gray-600">{screenshot.description}</p>
						</div>
						<div class="relative">
							<img 
								src={screenshot.src} 
								alt={screenshot.alt} 
								class="w-full rounded-xl shadow-lg border border-gray-200 hover:shadow-xl transition-shadow duration-300"
							/>
							<div class="absolute top-4 right-4">
								<Badge variant="secondary" class="bg-white/90 backdrop-blur-sm">
									<Calendar class="w-3 h-3 mr-1" />
									{screenshot.title.split(' ')[0]}
								</Badge>
							</div>
						</div>
					</div>
				{/each}
			</div>
		</div>
	</section>



	<!-- Installation Section -->
	<section id="installation" class="py-24 bg-white">
		<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
			<div class="text-center mb-16">
				<h2 class="text-4xl font-bold text-gray-900 mb-4">
					Get Started with Zendo
				</h2>
				<p class="text-xl text-gray-600 max-w-3xl mx-auto">
					Deploy Zendo quickly and easily using Docker. Get your personal task management system 
					running in minutes.
				</p>
			</div>

			<div class="grid lg:grid-cols-2 gap-12 items-start">
				<div class="space-y-8">
					<div class="space-y-6">
						<div class="flex items-center space-x-3">
							<div class="w-10 h-10 bg-gradient-to-br from-blue-100 to-indigo-100 rounded-xl flex items-center justify-center">
								<Dock class="w-5 h-5 text-blue-600" />
							</div>
							<div>
								<h3 class="text-lg font-semibold text-gray-900">Docker Installation</h3>
								<p class="text-gray-600">The fastest way to get Zendo running</p>
							</div>
						</div>

						<div class="space-y-4">
							<h4 class="text-md font-semibold text-gray-900">Prerequisites</h4>
							<ul class="space-y-2 text-gray-600">
								<li class="flex items-center space-x-2">
									<CheckCircle class="w-4 h-4 text-green-500" />
									<span>Docker and Docker Compose installed</span>
								</li>
								<li class="flex items-center space-x-2">
									<CheckCircle class="w-4 h-4 text-green-500" />
									<span>Port 8080 available</span>
								</li>
							</ul>
						</div>

						<div class="space-y-4">
							<h4 class="text-md font-semibold text-gray-900">Quick Start</h4>
							<div class="bg-gray-900 rounded-lg p-4 text-sm">
								<div class="text-green-400 mb-2"># Clone the repository</div>
								<div class="text-white">git clone https://github.com/yourusername/zendo.git</div>
								<div class="text-white">cd zendo/zendo-backend</div>
								<div class="text-green-400 mt-4 mb-2"># Start Zendo</div>
								<div class="text-white">docker-compose up -d</div>
								<div class="text-green-400 mt-4 mb-2"># Access the application</div>
								<div class="text-white">open http://localhost:8080</div>
							</div>
						</div>
					</div>
				</div>

				<div class="space-y-6">
					<div class="bg-gray-50 rounded-xl p-6 border border-gray-200">
						<h4 class="text-lg font-semibold text-gray-900 mb-4 flex items-center">
							<Terminal class="w-5 h-5 mr-2" />
							docker-compose.yaml
						</h4>
						<pre class="text-sm text-gray-700 overflow-x-auto"><code>{dockerComposeYaml}</code></pre>
					</div>

					<div class="space-y-4">
						<h4 class="text-lg font-semibold text-gray-900">Configuration Options</h4>
						<div class="space-y-3">
							<div class="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
								<span class="text-gray-700">Port</span>
								<Badge variant="secondary">8080</Badge>
							</div>
							<div class="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
								<span class="text-gray-700">Timezone</span>
								<Badge variant="secondary">America/Los_Angeles</Badge>
							</div>
							<div class="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
								<span class="text-gray-700">Storage</span>
								<Badge variant="secondary">./storage</Badge>
							</div>
						</div>
					</div>

					<!-- <div class="bg-blue-50 border border-blue-200 rounded-lg p-4">
						<h4 class="text-blue-900 font-semibold mb-2">💡 Pro Tip</h4>
						<p class="text-blue-800 text-sm">
							For ARM-based systems (like Apple Silicon Macs), change the platform to 
							<code class="bg-blue-100 px-1 rounded">linux/arm64</code> in the docker-compose.yaml file.
						</p>
					</div> -->
				</div>
			</div>
		</div>
	</section>
</div>
